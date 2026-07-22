<script setup lang="ts">
import { onBeforeUnmount, onMounted, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import FileResults from '../components/FileResults.vue'
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
    <form class="search-form" role="search" @submit.prevent="submit">
      <label class="visually-hidden" for="file-search">Search by tags</label>
      <input id="file-search" v-model="input" type="search" placeholder="rating:safe character:konata_izumi" autocomplete="off">
      <button class="button button--primary" type="submit">Search</button>
    </form>
    <p v-if="queryError" class="query-error" role="alert">{{ queryError }}</p>
    <details class="search-help">
      <summary>Search help</summary>
      <div><code>reviewed</code><span>has a boolean tag</span><code>!artist</code><span>does not have a tag</span><code>rating:safe</code><span>equals a value</span><code>score&gt;=10</code><span>number/date comparison</span><code>title:"hello world"</code><span>value containing spaces</span></div>
      <p>Conditions are combined with AND. OR and grouped expressions are not supported.</p>
    </details>
    <p v-if="!route.query.q" class="search-prompt">Enter one or more tag conditions.</p>
    <FileResults v-else :page="page" :collection="collection" :loading="loading" :error-message="errorMessage" empty-message="No files match this search." :return-to="route.fullPath" @page="run" @retry="run(page?.offset ?? 0)" />
  </main>
</template>

<style scoped>
.search-page { max-width: 90rem; }
.search-form { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: .55rem; max-width: 52rem; margin-top: 1.5rem; }
.search-form input { min-width: 0; padding: .75rem .85rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface); }
.query-error { color: var(--danger); }
.search-help { max-width: 52rem; margin-top: .75rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface); }
.search-help summary { min-height: 2.75rem; padding: .75rem; color: var(--primary); cursor: pointer; font-weight: 800; }
.search-help div { display: grid; grid-template-columns: max-content 1fr; gap: .5rem 1rem; padding: .25rem .75rem .75rem; }
.search-help code { color: var(--action); }
.search-help span, .search-help p, .search-prompt { color: var(--text-muted); }
.search-help p { margin: 0; padding: .75rem; border-top: 1px solid var(--border); font-size: .85rem; }
.search-prompt { margin-top: 3rem; }
@media (max-width: 30rem) { .search-form { grid-template-columns: 1fr; } .search-help div { grid-template-columns: 1fr; } }
</style>
