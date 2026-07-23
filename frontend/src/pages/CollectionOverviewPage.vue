<script setup lang="ts">
import { onMounted, ref, shallowRef, watch } from 'vue'

import { getCollection } from '../api'
import type { CollectionInfo } from '../api'

const props = defineProps<{ collection: string }>()
const info = ref<CollectionInfo>()
const loading = shallowRef(true)
const errorMessage = shallowRef('')

async function load() {
  loading.value = true
  try {
    info.value = await getCollection(props.collection)
    errorMessage.value = ''
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Collection is unavailable'
  } finally {
    loading.value = false
  }
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 ** 3) return `${(bytes / 1024 ** 2).toFixed(1)} MB`
  return `${(bytes / 1024 ** 3).toFixed(1)} GB`
}

watch(() => props.collection, () => void load())
onMounted(() => void load())
</script>

<template>
  <main class="page overview-page">
    <header class="page-heading"><p>Collection</p><h1>{{ collection }}</h1><span v-if="info?.is_default">Default collection</span></header>
    <p v-if="loading" class="empty-copy" role="status">Loading collection…</p>
    <div v-else-if="errorMessage" class="error-state" role="alert"><p>{{ errorMessage }}</p><button class="button" type="button" @click="load">Retry</button></div>
    <template v-else-if="info">
      <p v-if="info.comment" class="collection-comment">{{ info.comment }}</p>
      <dl class="collection-stats">
        <div><RouterLink :to="`/collections/${encodeURIComponent(collection)}/files`"><dt>Files</dt><dd>{{ info.file_count }}</dd></RouterLink></div>
        <div><RouterLink :to="`/collections/${encodeURIComponent(collection)}/files`"><dt>Indexed size</dt><dd>{{ formatBytes(info.total_size_bytes) }}</dd></RouterLink></div>
        <div><RouterLink :to="`/collections/${encodeURIComponent(collection)}/tags`"><dt>Tags</dt><dd>{{ info.tag_count }}</dd></RouterLink></div>
        <div><RouterLink :to="{ path: `/collections/${encodeURIComponent(collection)}/tags`, query: { required: 'true' } }"><dt>Required</dt><dd>{{ info.required_count }}</dd></RouterLink></div>
      </dl>
      <RouterLink class="storage-section" :to="`/collections/${encodeURIComponent(collection)}/storage`"><h2>Storage</h2><span v-for="storage in info.storages" :key="storage">{{ storage }}</span></RouterLink>
    </template>
  </main>
</template>

<style scoped>
.overview-page { display: grid; align-content: start; gap: 1.5rem; }
.page-heading span { display: inline-block; margin-top: .7rem; color: var(--primary); font-size: .82rem; font-weight: 800; }
.collection-comment { max-width: 60ch; margin: 0; color: var(--text-muted); line-height: 1.6; }
.collection-stats { display: grid; grid-template-columns: repeat(2, 1fr); gap: .65rem; margin: 0; }
.collection-stats div { overflow: hidden; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); }.collection-stats a { display: block; height: 100%; padding: 1rem; color: inherit; text-decoration: none; }.collection-stats a:hover { background: var(--primary-soft); }
dt { color: var(--text-muted); font-size: .78rem; font-weight: 800; text-transform: uppercase; }
dd { margin: .35rem 0 0; color: var(--primary); font: 1.8rem var(--display); }
.storage-section { display: flex; min-height: 4rem; flex-wrap: wrap; align-items: center; gap: .6rem; padding: .75rem; border: 1px solid var(--border); border-radius: var(--radius-lg); color: inherit; text-decoration: none; background: var(--surface); }
.storage-section h2 { width: 100%; margin: 0; font: 1.15rem var(--display); }
.storage-section span { padding: .4rem .65rem; border-radius: 999px; color: var(--success); background: color-mix(in srgb, var(--success) 12%, transparent); font-weight: 800; }
.error-state { color: var(--danger); }
@media (min-width: 48rem) { .collection-stats { grid-template-columns: repeat(4, 1fr); } }
</style>
