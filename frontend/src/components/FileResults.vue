<script setup lang="ts">
import type { FilePage } from '../api'
import FileCard from './FileCard.vue'

defineProps<{ page?: FilePage; collection: string; loading: boolean; errorMessage: string; emptyMessage: string; returnTo?: string }>()
const emit = defineEmits<{ page: [offset: number]; retry: [] }>()
</script>

<template>
  <div class="results" :aria-busy="loading">
    <p v-if="loading && !page" class="result-state" role="status">Loading files…</p>
    <div v-else-if="errorMessage" class="result-state" role="alert"><p>{{ errorMessage }}</p><button class="button" type="button" @click="emit('retry')">Retry</button></div>
    <p v-else-if="page && !page.files.length" class="result-state">{{ emptyMessage }}</p>
    <div v-else-if="page" class="file-grid"><FileCard v-for="file in page.files" :key="file.sha256" :file="file" :collection="collection" :return-to="returnTo" /></div>
    <nav v-if="page && (page.offset > 0 || page.has_more)" aria-label="Result pages">
      <button class="button" type="button" :disabled="loading || page.offset === 0" @click="emit('page', Math.max(0, page.offset - page.limit))">← Previous</button>
      <span>{{ page.offset + 1 }}–{{ page.offset + page.files.length }}</span>
      <button class="button" type="button" :disabled="loading || !page.has_more" @click="emit('page', page.next_offset ?? page.offset)">Next →</button>
    </nav>
  </div>
</template>

<style scoped>
.results { margin-top: 1.5rem; }
.file-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: .65rem; opacity: 1; transition: opacity .12s ease; }
[aria-busy='true'] .file-grid { opacity: .55; }
.result-state { max-width: 36rem; margin: 3rem 0; color: var(--text-muted); }
nav { display: flex; align-items: center; justify-content: space-between; gap: .75rem; margin-top: 1.5rem; }
nav span { color: var(--text-muted); font-size: .85rem; }
@media (min-width: 40rem) { .file-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 1rem; } }
@media (min-width: 64rem) { .file-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); } }
@media (min-width: 90rem) { .file-grid { grid-template-columns: repeat(5, minmax(0, 1fr)); } }
</style>
