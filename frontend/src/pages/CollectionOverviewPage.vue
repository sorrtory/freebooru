<script setup lang="ts">
defineProps<{ collection: string }>()
</script>

<template>
  <main class="page overview-page">
    <header class="page-heading"><p>Collection</p><h1>{{ collection }}</h1></header>
    <div class="quick-actions">
      <RouterLink :to="`/collections/${encodeURIComponent(collection)}/files`"><strong>All files</strong><span>Browse the collection</span></RouterLink>
      <RouterLink :to="`/collections/${encodeURIComponent(collection)}/search`"><strong>Search</strong><span>Find files by tags</span></RouterLink>
      <RouterLink class="accent" :to="`/collections/${encodeURIComponent(collection)}/import`"><strong>Upload</strong><span>Add one file</span></RouterLink>
    </div>
  </main>
</template>

<style scoped>
.overview-page { display: grid; align-content: start; gap: 2rem; }
.quick-actions { display: grid; gap: .75rem; }
.quick-actions a { display: grid; min-height: 6.5rem; align-content: center; gap: .35rem; padding: 1.25rem; border: 1px solid var(--border); border-radius: var(--radius-lg); color: var(--text); background: var(--surface); text-decoration: none; transition: transform .15s ease, border-color .15s ease; }
.quick-actions a:hover { border-color: var(--primary); transform: translateY(-2px); }
.quick-actions strong { color: var(--primary); font: 1.25rem var(--display); }
.quick-actions span { color: var(--text-muted); }
.quick-actions .accent { border-color: color-mix(in srgb, var(--action) 45%, var(--border)); }
.quick-actions .accent strong { color: var(--action); }
@media (min-width: 48rem) { .quick-actions { grid-template-columns: repeat(3, 1fr); } }
</style>
