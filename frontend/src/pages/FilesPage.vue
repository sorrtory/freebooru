<script setup lang="ts">
import { onBeforeUnmount, onMounted, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import FileResults from '../components/FileResults.vue'
import SearchSidebar from '../components/SearchSidebar.vue'
import { parseSearchQuery } from '../searchQuery'
import { useFileSearch } from '../useFileSearch'

const props = defineProps<{ collection: string }>()
const route = useRoute()
const router = useRouter()
const input = shallowRef(String(route.query.q ?? ''))
const queryError = shallowRef('')
const { page, loading, errorMessage, load, cancel } = useFileSearch(props.collection)

function loadPage(offset = 0) {
  try {
    queryError.value = ''
    void load(parseSearchQuery(input.value), offset)
  } catch (error) {
    queryError.value = error instanceof Error ? error.message : 'Invalid search'
  }
}

async function syncQuery() {
  try {
    parseSearchQuery(input.value)
    queryError.value = ''
    const q = input.value.trim()
    await router.replace({ query: q ? { q } : {} })
    loadPage()
  } catch (error) {
    queryError.value = error instanceof Error ? error.message : 'Invalid search'
  }
}

watch(() => route.query.q, (value) => {
  const next = String(value ?? '')
  if (input.value !== next) input.value = next
  loadPage()
})
onMounted(loadPage)
onBeforeUnmount(cancel)
</script>

<template>
  <main class="page files-page">
    <header class="page-heading"><p>{{ collection }}</p><h1>Files</h1></header>
    <div class="files-layout">
      <SearchSidebar v-model="input" :collection="collection" :error-message="queryError" @change="syncQuery" @search="syncQuery" />
      <FileResults :page="page" :collection="collection" :loading="loading" :error-message="errorMessage" :empty-message="input ? 'No files match this search.' : 'This collection has no files yet.'" :return-to="route.fullPath" @page="loadPage" @retry="loadPage(page?.offset ?? 0)" />
    </div>
  </main>
</template>

<style scoped>
.files-page { max-width: 94rem; }.files-layout { display: grid; gap: 1rem; margin-top: 1.5rem; }.files-layout > :first-child { padding: 1rem; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); }
@media (min-width: 64rem) { .files-layout { grid-template-columns: clamp(16rem, 23vw, 23rem) minmax(0, 1fr); align-items: start; }.files-layout > :first-child { position: sticky; top: 5rem; } }
</style>
