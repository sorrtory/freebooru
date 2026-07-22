import { ref, shallowRef } from 'vue'

import { getFiles } from './api'
import type { FilePage } from './api'

export function useFileSearch(collection: string) {
  const page = ref<FilePage>()
  const loading = shallowRef(false)
  const errorMessage = shallowRef('')
  let request: AbortController | undefined

  async function load(terms: string[], offset = 0) {
    request?.abort()
    const controller = new AbortController()
    request = controller
    loading.value = true
    try {
      const result = await getFiles(collection, terms, offset, controller.signal)
      if (request !== controller) return
      page.value = result
      errorMessage.value = ''
    } catch (error) {
      if (controller.signal.aborted) return
      errorMessage.value = error instanceof Error ? error.message : 'Files are unavailable'
    } finally {
      if (request === controller) loading.value = false
    }
  }

  function cancel() { request?.abort() }
  return { page, loading, errorMessage, load, cancel }
}
