<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, shallowRef, useTemplateRef } from 'vue'
import { useRouter } from 'vue-router'

import { useCollections } from '../useCollections'

const props = defineProps<{ collection: string }>()
const router = useRouter()
const { collections, loading, errorMessage, load } = useCollections()
const open = shallowRef(false)
const query = shallowRef('')
const activeIndex = shallowRef(0)
const search = useTemplateRef<HTMLInputElement>('search')
const root = useTemplateRef<HTMLElement>('root')
const filtered = computed(() => {
  const normalized = query.value.trim().toLocaleLowerCase()
  return collections.value.filter((item) => item.name.toLocaleLowerCase().includes(normalized))
})

function toggle() {
  open.value = !open.value
  if (open.value) {
    query.value = ''
    activeIndex.value = Math.max(0, filtered.value.findIndex((item) => item.name === props.collection))
    requestAnimationFrame(() => search.value?.focus())
  }
}

function move(direction: number) {
  activeIndex.value = Math.max(0, Math.min(filtered.value.length, activeIndex.value + direction))
}

function choose(name: string) {
  open.value = false
  void router.push(`/collections/${encodeURIComponent(name)}`)
}

function commitActive() {
  const item = filtered.value[activeIndex.value]
  if (item) choose(item.name)
  else void router.push('/collections?create=1')
}

function outside(event: PointerEvent) { if (!root.value?.contains(event.target as Node)) open.value = false }
onMounted(() => { document.addEventListener('pointerdown', outside); void load() })
onBeforeUnmount(() => document.removeEventListener('pointerdown', outside))
</script>

<template>
  <div ref="root" class="switcher">
    <button class="switcher-button" type="button" :aria-expanded="open" aria-haspopup="listbox" @click="toggle">
      <strong>{{ collection }}</strong><span aria-hidden="true">⌄</span>
    </button>
    <div v-if="open" class="switcher-popover">
      <input ref="search" v-model="query" role="combobox" aria-label="Find collection" :aria-expanded="open" aria-controls="collection-options" :aria-activedescendant="`collection-option-${activeIndex}`" placeholder="Find collection…" @keydown.down.prevent="move(1)" @keydown.up.prevent="move(-1)" @keydown.enter.prevent="commitActive" @keydown.esc="open = false">
      <div id="collection-options" role="listbox">
        <button v-for="(item, index) in filtered" :id="`collection-option-${index}`" :key="item.name" type="button" role="option" :aria-selected="item.name === collection" :class="{ active: index === activeIndex }" @mouseenter="activeIndex = index" @click="choose(item.name)">
          <span>{{ item.name }}</span><small v-if="item.is_default">default</small>
        </button>
        <RouterLink :id="`collection-option-${filtered.length}`" role="option" class="add-collection" :class="{ active: activeIndex === filtered.length }" to="/collections?create=1" @click="open = false">+ Add collection…</RouterLink>
      </div>
      <p v-if="loading">Loading…</p><p v-else-if="errorMessage" role="alert">{{ errorMessage }}</p>
    </div>
  </div>
</template>

<style scoped>
.switcher { position: relative; min-width: 0; }
.switcher-button { display: flex; width: 100%; min-width: 0; min-height: 2.75rem; align-items: center; justify-content: space-between; gap: .5rem; padding: .4rem .7rem; border: 1px solid var(--border); border-radius: var(--radius); color: var(--text); background: var(--surface-raised); cursor: pointer; }
.switcher-button strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.switcher-button span { color: var(--primary); }
.switcher-popover { position: absolute; top: calc(100% + .5rem); left: 0; z-index: 40; width: min(21rem, calc(100vw - 2rem)); padding: .5rem; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); box-shadow: var(--shadow); }
input { width: 100%; padding: .65rem .75rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--canvas); }
[role='listbox'] { display: grid; max-height: min(22rem, 55vh); margin-top: .4rem; overflow: auto; }
[role='option'], .add-collection { display: flex; min-height: 2.75rem; align-items: center; justify-content: space-between; padding: .55rem .65rem; border: 0; border-radius: var(--radius); color: var(--text); background: transparent; cursor: pointer; text-align: left; text-decoration: none; }
[role='option'].active, [role='option']:hover, .add-collection.active, .add-collection:hover { background: var(--primary-soft); }
[role='option'][aria-selected='true'] span { color: var(--primary); font-weight: 800; }
small { color: var(--text-muted); }
.add-collection { margin-top: .35rem; border-top: 1px solid var(--border); border-radius: 0 0 var(--radius) var(--radius); color: var(--action); font-weight: 800; }
.switcher-popover p { margin: .5rem; color: var(--text-muted); font-size: .85rem; }
</style>
