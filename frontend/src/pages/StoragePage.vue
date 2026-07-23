<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue'
import { getCollectionStorages, importCollectionStorage, type CollectionStorage } from '../api'

const props = defineProps<{ collection: string }>()
const storages = shallowRef<CollectionStorage[]>([])
const error = shallowRef('')
const loading = shallowRef(true)
const importing = shallowRef('')
const showOther = shallowRef(false)
const imported = computed(() => storages.value.filter((item) => item.imported))
const available = computed(() => storages.value.filter((item) => !item.imported))
onMounted(load)
async function load() { loading.value = true; try { storages.value = await getCollectionStorages(props.collection); error.value = '' } catch (cause) { error.value = cause instanceof Error ? cause.message : 'Storage is unavailable.' } finally { loading.value = false } }
async function add(item: CollectionStorage) { importing.value = item.name; try { await importCollectionStorage(props.collection, item.name); await load() } catch (cause) { error.value = cause instanceof Error ? cause.message : 'Storage could not be imported.' } finally { importing.value = '' } }
function size(bytes: number) { return new Intl.NumberFormat(undefined, { notation: 'compact', style: 'unit', unit: 'byte' }).format(bytes) }
</script>

<template>
  <main class="page storage-page"><header class="page-heading"><p>{{ collection }}</p><h1>Storage</h1></header><p v-if="error" class="error" role="alert">{{ error }}</p><p v-if="loading" role="status">Loading storage…</p><template v-else><section><h2>In this collection</h2><details v-for="item in imported" :key="item.name"><summary><span><strong>{{ item.name }}</strong><small>{{ item.type }}</small></span><span>{{ item.file_count }} files · {{ size(item.total_size_bytes) }}</span></summary><p>{{ item.comment || 'No description.' }}</p></details></section><section v-if="available.length"><button class="other-toggle" type="button" :aria-expanded="showOther" @click="showOther = !showOther">Other configured storage <span>{{ available.length }} {{ showOther ? '−' : '+' }}</span></button><article v-for="item in showOther ? available : []" :key="item.name"><div><strong>{{ item.name }}</strong><p>{{ item.comment || item.type }}</p></div><button class="button" :disabled="Boolean(importing)" @click="add(item)">{{ importing === item.name ? 'Importing…' : 'Import' }}</button></article></section></template></main>
</template>

<style scoped>
.storage-page { display: grid; align-content: start; gap: 1.25rem; }.error { color: var(--danger); }section { overflow: hidden; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); }section h2, .other-toggle { margin: 0; padding: 1rem; font: 1.5rem var(--display); }.other-toggle { display: flex; width: 100%; min-height: 4rem; align-items: center; justify-content: space-between; border: 0; color: var(--text); background: var(--surface-raised); text-align: left; }details, article { border-top: 1px solid var(--line); }summary, article { display: flex; min-height: 4rem; align-items: center; justify-content: space-between; gap: 1rem; padding: .75rem 1rem; }summary { cursor: pointer; }summary span:first-child { display: grid; }small, summary span:last-child, article p { color: var(--text-muted); }details > p { margin: 0; padding: 0 1rem 1rem; }article p { margin: .2rem 0 0; }
</style>
