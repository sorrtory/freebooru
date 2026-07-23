<script setup lang="ts">
import { computed, onMounted, shallowRef, watch } from 'vue'

import { getCollectionTags, type CollectionTag } from '../api'
import { parseSearchQuery } from '../searchQuery'

const props = defineProps<{ collection: string; modelValue: string; errorMessage: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string]; search: []; change: [] }>()
const input = shallowRef(props.modelValue)
const tags = shallowRef<CollectionTag[]>([])
const open = shallowRef(false)
const active = shallowRef(0)

watch(() => props.modelValue, (value) => { input.value = value })
watch(input, (value, _previous, onCleanup) => {
  emit('update:modelValue', value)
  const timer = window.setTimeout(() => emit('change'), 180)
  onCleanup(() => window.clearTimeout(timer))
})
onMounted(async () => { try { tags.value = (await getCollectionTags(props.collection)).filter((tag) => tag.imported && !tag.system) } catch { tags.value = [] } })

const hot = computed(() => [...tags.value].filter((tag) => tag.assignment_count > 0).sort((left, right) => right.assignment_count - left.assignment_count || left.name.localeCompare(right.name)).slice(0, 12))
const terms = computed(() => { try { return parseSearchQuery(input.value) } catch { return [] } })
const suggestions = computed(() => {
  const token = input.value.slice(input.value.lastIndexOf(' ') + 1).toLowerCase()
  if (!token) return []
  const colon = token.indexOf(':')
  if (colon >= 0) {
    const name = token.slice(0, colon)
    const valuePrefix = token.slice(colon + 1)
    const tag = tags.value.find((item) => item.name.toLowerCase() === name)
    return tag?.values.filter((value) => value.value.toLowerCase().startsWith(valuePrefix)).map((value) => `${tag.name}:${quote(value.value)}`).slice(0, 8) ?? []
  }
  return tags.value.filter((tag) => tag.name.toLowerCase().startsWith(token)).map((tag) => tag.type === 'bool' ? tag.name : `${tag.name}:`).slice(0, 8)
})

function quote(value: string) { return /\s/.test(value) ? `"${value}"` : value }
function replaceToken(value: string) { const boundary = input.value.lastIndexOf(' '); input.value = `${boundary >= 0 ? input.value.slice(0, boundary + 1) : ''}${value}`; active.value = 0; open.value = value.endsWith(':') }
function complete(event: KeyboardEvent) { if (!suggestions.value.length) return; event.preventDefault(); replaceToken(suggestions.value[active.value] ?? suggestions.value[0]) }
function move(step: number) { if (!suggestions.value.length) return; open.value = true; active.value = (active.value + step + suggestions.value.length) % suggestions.value.length }
function addHot(tag: CollectionTag) { const next = [...terms.value, tag.name]; input.value = next.join(' '); emit('search') }
function removeTerm(index: number) { input.value = terms.value.filter((_, current) => current !== index).map(quote).join(' '); emit('search') }
</script>

<template>
  <aside class="search-sidebar">
    <form role="search" @submit.prevent="emit('search')">
      <label for="file-search">Search</label>
      <div class="search-control"><svg aria-hidden="true" viewBox="0 0 24 24"><circle cx="11" cy="11" r="6"/><path d="m16 16 4 4"/></svg><input id="file-search" v-model="input" type="search" placeholder="rating:safe" autocomplete="off" role="combobox" :aria-expanded="open && Boolean(suggestions.length)" @focus="open = true" @input="open = true; active = 0" @keydown.tab="complete" @keydown.down.prevent="move(1)" @keydown.up.prevent="move(-1)" @keydown.escape="open = false"></div>
      <div v-if="open && suggestions.length" class="suggestions" role="listbox"><button v-for="(suggestion, index) in suggestions" :key="suggestion" type="button" role="option" :aria-selected="active === index" @mousedown.prevent="replaceToken(suggestion)">{{ suggestion }}</button></div>
      <button class="button button--primary" type="submit">Search</button>
    </form>
    <p v-if="errorMessage" class="query-error" role="alert">{{ errorMessage }}</p>
    <div v-if="terms.length" class="active-terms"><button v-for="(term, index) in terms" :key="`${term}-${index}`" type="button" :aria-label="`Remove ${term}`" @click="removeTerm(index)">{{ term }} ×</button></div>
    <section v-if="hot.length" aria-labelledby="hot-tags"><h2 id="hot-tags">Hot tags</h2><button v-for="tag in hot" :key="tag.name" type="button" @click="addHot(tag)"><span>{{ tag.name }}</span><small>{{ tag.assignment_count }}</small></button></section>
    <details><summary>Search help</summary><p><code>tag</code> present</p><p><code>!tag</code> absent</p><p><code>rating:safe</code> equals</p><p><code>score&gt;=10</code> comparison</p><p>Use quotes around values with spaces. Conditions use AND.</p></details>
  </aside>
</template>

<style scoped>
.search-sidebar { position: relative; }.search-sidebar form { display: grid; position: relative; gap: .5rem; }.search-sidebar form > label, section h2 { font-size: .76rem; font-weight: 800; letter-spacing: .07em; text-transform: uppercase; }.search-control { display: flex; align-items: center; border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface); }.search-control:focus-within { border-color: var(--primary); }.search-control svg { width: 1.1rem; margin-left: .7rem; fill: none; stroke: var(--text-muted); stroke-width: 2; }.search-control input { width: 100%; min-width: 0; padding: .65rem; border: 0; outline: 0; background: transparent; }.suggestions { position: absolute; z-index: 10; top: 5rem; width: 100%; padding: .25rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface); box-shadow: var(--shadow); }.suggestions button, section button { display: flex; width: 100%; min-height: 2.75rem; align-items: center; justify-content: space-between; padding: .5rem .65rem; border: 0; border-radius: .25rem; color: var(--text); background: transparent; text-align: left; }.suggestions button[aria-selected='true'], .suggestions button:hover, section button:hover { background: var(--primary-soft); }.active-terms { display: flex; flex-wrap: wrap; gap: .35rem; margin-top: .8rem; }.active-terms button { min-height: 2.25rem; padding: .3rem .55rem; border: 1px solid var(--border); border-radius: 999px; color: var(--primary); background: var(--primary-soft); }.search-sidebar section { margin-top: 1.25rem; }.search-sidebar section h2 { margin: 0 0 .35rem; }.search-sidebar section small { color: var(--text-muted); }.search-sidebar details { margin-top: 1rem; border-top: 1px solid var(--border); }.search-sidebar summary { min-height: 2.75rem; padding: .75rem 0; color: var(--primary); cursor: pointer; font-weight: 800; }.search-sidebar details p { margin: .4rem 0; color: var(--text-muted); font-size: .82rem; }.query-error { color: var(--danger); }
</style>
