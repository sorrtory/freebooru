<script setup lang="ts">
import { computed } from 'vue'

import type { FileRecord } from '../api'

const props = defineProps<{ file: FileRecord; collection: string; returnTo?: string }>()
const visual = computed(() => props.file.mime_type.startsWith('image/'))
const visibleAssignments = computed(() => props.file.assignments.slice(0, 3))
const hiddenAssignments = computed(() => Math.max(0, props.file.assignments.length - visibleAssignments.value.length))
const route = computed(() => ({
  path: `/collections/${encodeURIComponent(props.collection)}/files/${props.file.sha256}`,
  query: props.returnTo ? { from: props.returnTo } : undefined,
}))

function formatBytes(bytes: number) {
  if (bytes < 1024 ** 2) return `${Math.max(1, Math.round(bytes / 1024))} KB`
  return `${(bytes / 1024 ** 2).toFixed(1)} MB`
}
function formatAssignment(assignment: FileRecord['assignments'][number]) {
  if (assignment.value === true) return assignment.name
  const value = Array.isArray(assignment.value) ? assignment.value.join(', ') : String(assignment.value)
  return `${assignment.name}:${value}`
}
</script>

<template>
  <RouterLink class="file-card" :to="route">
    <div class="preview">
      <img v-if="visual" :src="file.content_url" :alt="file.filename" loading="lazy">
      <span v-else aria-hidden="true">{{ file.mime_type.split('/')[1]?.slice(0, 5).toUpperCase() || 'FILE' }}</span>
    </div>
    <div class="file-copy"><strong :title="file.filename">{{ file.filename }}</strong><small>{{ formatBytes(file.size_bytes) }}</small><div v-if="visibleAssignments.length" class="file-tags"><span v-for="assignment in visibleAssignments" :key="assignment.name" :title="formatAssignment(assignment)">{{ formatAssignment(assignment) }}</span><span v-if="hiddenAssignments" :title="`${file.assignments.length} tags total`">+{{ hiddenAssignments }}</span></div></div>
  </RouterLink>
</template>

<style scoped>
.file-card { display: grid; min-width: 0; overflow: hidden; border: 1px solid var(--border); border-radius: var(--radius-lg); color: var(--text); background: var(--surface); text-decoration: none; transition: transform .14s ease, border-color .14s ease, box-shadow .14s ease; }
.file-card:hover { border-color: var(--primary); box-shadow: var(--shadow); transform: translateY(-2px); }
.preview { display: grid; aspect-ratio: 1; place-items: center; overflow: hidden; color: var(--primary); background: linear-gradient(145deg, var(--primary-soft), var(--surface-raised)); font: 800 .8rem var(--mono); }
.preview img { width: 100%; height: 100%; object-fit: cover; }
.file-copy { display: grid; min-width: 0; gap: .25rem; padding: .75rem; }
.file-copy strong { overflow: hidden; font-size: .9rem; text-overflow: ellipsis; white-space: nowrap; }
.file-copy small { color: var(--text-muted); font-size: .75rem; }
.file-tags { display: flex; min-width: 0; flex-wrap: wrap; gap: .25rem; }.file-tags span { overflow: hidden; max-width: 100%; padding: .18rem .4rem; border-radius: 999px; color: var(--primary); background: var(--primary-soft); font-size: .68rem; font-weight: 750; text-overflow: ellipsis; white-space: nowrap; }
</style>
