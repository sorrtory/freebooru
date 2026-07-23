<script setup lang="ts">
import { onBeforeUnmount, onMounted, shallowRef, useTemplateRef } from 'vue'

import type { CollectionTag } from '../api'
import { parseSearchQuery } from '../searchQuery'
import { quoteSearchValue, replaceSearchToken, useSearchHints } from '../useSearchHints'

const props = defineProps<{ collection: string }>()
const model = defineModel<string>({ required: true })
const emit = defineEmits<{ submit: [query: string] }>()
const root = useTemplateRef<HTMLElement>('root')
const open = shallowRef(false)
const { suggestions: completions, popular } = useSearchHints(props.collection, model)

function replaceToken(value: string) { model.value = replaceSearchToken(model.value, value); open.value = true }
function choosePopular(tag: CollectionTag) { const terms = parseSearchQuery(model.value); const next = terms.includes(tag.name) ? terms.filter((term) => term !== tag.name) : [...terms, tag.name]; model.value = next.map(quoteSearchValue).join(' '); emit('submit', model.value); open.value = false }
function submit() { emit('submit', model.value); open.value = false }
function outside(event: PointerEvent) { if (!root.value?.contains(event.target as Node)) open.value = false }
onMounted(() => document.addEventListener('pointerdown', outside))
onBeforeUnmount(() => document.removeEventListener('pointerdown', outside))
</script>

<template>
  <form ref="root" class="header-search" role="search" @submit.prevent="submit">
    <div class="header-search-control"><svg aria-hidden="true" viewBox="0 0 24 24"><circle cx="11" cy="11" r="6"/><path d="m16 16 4 4"/></svg><input v-model="model" type="search" aria-label="Search files" placeholder="Search tags" autocomplete="off" @focus="open = true" @input="open = true"></div>
    <div v-if="open && (completions.length || popular.length)" class="search-popover">
      <section v-if="completions.length"><h2>Complete</h2><button v-for="item in completions" :key="item" type="button" @mousedown.prevent="replaceToken(item)">{{ item }}</button></section>
      <section v-if="popular.length"><h2>Popular</h2><button v-for="tag in popular" :key="tag.name" type="button" @mousedown.prevent="choosePopular(tag)"><span>{{ tag.name }}</span><small>{{ tag.assignment_count }}</small></button></section>
    </div>
  </form>
</template>

<style scoped>
.header-search { position: relative; min-width: 0; flex: 1 1 18rem; }.header-search-control { display: flex; min-height: 2.75rem; align-items: center; border: 1px solid var(--border); border-radius: var(--radius); background: var(--background); }.header-search-control:focus-within { border-color: var(--primary); box-shadow: 0 0 0 3px var(--primary-soft); }.header-search svg { width: 1.05rem; margin-left: .7rem; fill: none; stroke: var(--text-muted); stroke-width: 2; }.header-search input { width: 100%; min-width: 0; padding: .55rem .7rem; border: 0; outline: 0; color: var(--text); background: transparent; font: inherit; }
.search-popover { display: grid; position: absolute; z-index: 40; top: calc(100% + .4rem); width: min(32rem, calc(100vw - 1.5rem)); max-height: min(26rem, 70vh); overflow: auto; padding: .45rem; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); box-shadow: var(--shadow); }.search-popover h2 { margin: .45rem .55rem .25rem; color: var(--text-muted); font-size: .72rem; letter-spacing: .08em; text-transform: uppercase; }.search-popover button { display: flex; width: 100%; min-height: 2.75rem; align-items: center; justify-content: space-between; padding: .5rem .65rem; border: 0; border-radius: var(--radius); color: var(--text); background: transparent; text-align: left; }.search-popover button:hover { color: var(--primary); background: var(--primary-soft); }.search-popover small { color: var(--text-muted); }
</style>
