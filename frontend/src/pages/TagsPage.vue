<script setup lang="ts">
import { computed, onMounted, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getCollectionTags, importCollectionTag, type CollectionTag } from '../api'

const props = defineProps<{ collection: string }>()
const route = useRoute()
const router = useRouter()
const tags = shallowRef<CollectionTag[]>([])
const query = shallowRef(typeof route.query.q === 'string' ? route.query.q : '')
type TagMode = 'all' | 'required' | 'optional'
const mode = shallowRef<TagMode>(route.query.kind === 'optional' ? 'optional' : route.query.kind === 'required' || route.query.required === 'true' ? 'required' : 'all')
const loading = shallowRef(true)
const error = shallowRef('')
const importing = shallowRef('')
const showOther = shallowRef(false)
const filtered = computed(() => {
  const needle = query.value.trim().toLowerCase()
  return tags.value.filter((tag) => (!needle || tag.name.toLowerCase().includes(needle) || tag.comment.toLowerCase().includes(needle)) && (mode.value === 'all' || tag.required === (mode.value === 'required')))
})
const imported = computed(() => filtered.value.filter((tag) => tag.imported))
const available = computed(() => filtered.value.filter((tag) => !tag.imported))

onMounted(load)
watch([query, mode], () => {
  const kind = mode.value === 'all' ? undefined : mode.value
  void router.replace({ query: { ...(query.value ? { q: query.value } : {}), ...(kind ? { kind } : {}) } })
})

async function load() {
  loading.value = true; error.value = ''
  try { tags.value = await getCollectionTags(props.collection) }
  catch (cause) { error.value = cause instanceof Error ? cause.message : 'Tags are unavailable.' }
  finally { loading.value = false }
}

function toggle(next: Exclude<TagMode, 'all'>) { mode.value = mode.value === next ? 'all' : next }

async function add(tag: CollectionTag) {
  importing.value = tag.name; error.value = ''
  try { await importCollectionTag(props.collection, tag.name); await load() }
  catch (cause) { error.value = cause instanceof Error ? cause.message : 'Tag could not be imported.' }
  finally { importing.value = '' }
}
function searchTag(tag: CollectionTag) { void router.push({ path: `/collections/${encodeURIComponent(props.collection)}/files`, query: { q: tag.name } }) }
</script>

<template>
  <main class="page resource-page">
    <header class="page-heading"><p>{{ collection }}</p><h1>Tags</h1></header>
    <div class="resource-filter" role="search"><input v-model="query" type="search" aria-label="Find a tag" placeholder="Find a tag"><div class="kind-toggles" aria-label="Tag kind"><button type="button" :aria-pressed="mode === 'required'" @click="toggle('required')"><span aria-hidden="true">✓</span> Required</button><button type="button" :aria-pressed="mode === 'optional'" @click="toggle('optional')"><span aria-hidden="true">✓</span> Optional</button></div></div>
    <p v-if="error" class="resource-error" role="alert">{{ error }}</p><p v-if="loading" role="status">Loading tags…</p>
    <section v-else aria-labelledby="collection-tags"><header><h2 id="collection-tags">In this collection</h2><span>{{ imported.length }}</span></header><div class="resource-list"><article v-for="tag in imported" :key="tag.name" class="tag-card" role="link" tabindex="0" @click="searchTag(tag)" @keydown.enter="searchTag(tag)"><div><div class="tag-title"><strong>{{ tag.name }}</strong><span v-for="group in tag.groups ?? []" :key="group">{{ group }}</span></div><p>{{ tag.comment || tag.type }}</p></div><div class="resource-meta"><span v-if="tag.required">Required</span><span>{{ tag.assignment_count }} files</span><span v-if="tag.values.length">{{ tag.values.length }} values</span></div></article><p v-if="!imported.length">No matching imported tags.</p></div></section>
    <section v-if="available.length" aria-labelledby="other-tags"><button class="section-toggle" type="button" :aria-expanded="showOther" @click="showOther = !showOther"><span><small>Available</small><strong id="other-tags">Other configured tags</strong></span><b>{{ available.length }} {{ showOther ? '−' : '+' }}</b></button><div v-if="showOther" class="resource-list"><article v-for="tag in available" :key="tag.name"><div><div class="tag-title"><strong>{{ tag.name }}</strong><span v-for="group in tag.groups ?? []" :key="group">{{ group }}</span></div><p>{{ tag.comment || tag.type }}</p></div><button class="button" type="button" :disabled="Boolean(importing)" @click="add(tag)">{{ importing === tag.name ? 'Importing…' : 'Import' }}</button></article></div></section>
  </main>
</template>

<style scoped>
.resource-page { display: grid; align-content: start; gap: 1.25rem; }
.resource-filter { display: grid; gap: .5rem; padding: .75rem; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); }
.resource-filter > input { min-width: 0; padding: .65rem .75rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--canvas); }
.kind-toggles { display: flex; flex-wrap: wrap; gap: .4rem; }.kind-toggles button { min-height: 2.75rem; padding: .45rem .7rem; border: 1px solid var(--border); border-radius: var(--radius); color: var(--text-muted); background: var(--canvas); font-weight: 750; }.kind-toggles button span { opacity: 0; }.kind-toggles button[aria-pressed='true'] { border-color: var(--primary); color: var(--primary); background: var(--primary-soft); }.kind-toggles button[aria-pressed='true'] span { opacity: 1; }
section { overflow: hidden; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); }
section > header, .section-toggle { display: flex; width: 100%; min-height: 4rem; align-items: center; justify-content: space-between; padding: .75rem 1rem; border: 0; border-bottom: 1px solid var(--border); color: var(--text); background: var(--surface-raised); text-align: left; }
section h2 { margin: 0; font: 1.5rem var(--display); }
.section-toggle span { display: grid; gap: .15rem; }.section-toggle small { color: var(--text-muted); text-transform: uppercase; }.section-toggle strong { font-size: 1rem; }
.resource-list { display: grid; }.resource-list article { display: flex; min-height: 4rem; align-items: center; justify-content: space-between; gap: 1rem; padding: .75rem 1rem; border-top: 1px solid var(--line); }.resource-list article:first-child { border-top: 0; }.resource-list p { margin: .2rem 0 0; color: var(--text-muted); }.resource-meta { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: .35rem; }.resource-meta span { padding: .25rem .45rem; border-radius: 999px; color: var(--primary); background: var(--primary-soft); font-size: .72rem; font-weight: 800; }.resource-error { color: var(--danger); }
.tag-title { display: flex; flex-wrap: wrap; align-items: center; gap: .35rem; }.tag-title strong { color: var(--text); }.tag-title span { padding: .15rem .35rem; border: 1px solid var(--border); border-radius: 999px; color: var(--text-muted); font-size: .68rem; }.tag-card { cursor: pointer; }.tag-card:hover, .tag-card:focus-visible { outline: 0; background: var(--primary-soft); }
@media (min-width: 48rem) { .resource-filter { grid-template-columns: minmax(14rem, 1fr) auto; align-items: center; } }
</style>
