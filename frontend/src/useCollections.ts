import { readonly, ref, shallowRef } from 'vue'

import { createCollection as createCollectionRequest, getCollections } from './api'
import type { CollectionSummary } from './api'

const collections = ref<CollectionSummary[]>([])
const loaded = shallowRef(false)
const loading = shallowRef(false)
const errorMessage = shallowRef('')

export function useCollections() {
  async function load(force = false) {
    if (loading.value || (loaded.value && !force)) return
    loading.value = true
    try {
      const result = await getCollections()
      collections.value = result.collections
      loaded.value = true
      errorMessage.value = ''
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : 'Collections are unavailable'
    } finally {
      loading.value = false
    }
  }

  async function create(name: string) {
    const result = await createCollectionRequest(name)
    await load(true)
    return result
  }

  return { collections: readonly(collections), loaded: readonly(loaded), loading: readonly(loading), errorMessage: readonly(errorMessage), load, create }
}
