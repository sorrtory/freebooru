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

function currentTerms() {
  return parseSearchQuery(String(route.query.q ?? ''))
}

function run(offset = 0) {
  try {
    queryError.value = ''
    void load(currentTerms(), offset)
  } catch (error) {
    queryError.value = error instanceof Error ? error.message : 'Invalid search'
  }
}

async function submit() {
  try {
    parseSearchQuery(input.value)
    queryError.value = ''
    const next = input.value.trim()
    if (next === String(route.query.q ?? '')) run()
    else await router.push({ query: next ? { q: next } : {} })
  } catch (error) {
    queryError.value = error instanceof Error ? error.message : 'Invalid search'
  }
}

watch(() => route.query.q, (value) => { input.value = String(value ?? ''); run() })
onMounted(() => { if (route.query.q) run() })
onBeforeUnmount(cancel)
</script>

<template>
  <main class="page search-page">
    <header class="page-heading"><p>{{ collection }}</p><h1>Search</h1></header>
    <div class="search-layout"><SearchSidebar v-model="input" :collection="collection" :error-message="queryError" @search="submit" /><section class="search-results"><p v-if="!route.query.q" class="search-prompt">Choose a hot tag or enter one or more conditions.</p><FileResults v-else :page="page" :collection="collection" :loading="loading" :error-message="errorMessage" empty-message="No files match this search." :return-to="route.fullPath" @page="run" @retry="run(page?.offset ?? 0)" /></section></div>
  </main>
</template>

<style scoped>
.search-page { max-width: 90rem; }
.search-layout { display: grid; gap: 1rem; margin-top: 1.5rem; }.search-layout > :first-child { padding: 1rem; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); }.search-results { min-width: 0; }.search-prompt { color: var(--text-muted); }
.search-prompt { margin-top: 3rem; }
@media (min-width: 64rem) { .search-layout { grid-template-columns: minmax(14rem, 18rem) minmax(0, 1fr); align-items: start; }.search-layout > :first-child { position: sticky; top: 5rem; } }
</style>
