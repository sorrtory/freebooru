<script setup lang="ts">
defineProps<{ collection: string }>()
</script>

<template>
  <div class="app-shell">
    <header class="app-bar">
      <RouterLink class="brand" to="/">FreeBooru</RouterLink>
      <button class="collection-button" type="button" title="Choose collection">
        <span>Collection</span><strong>{{ collection }}</strong><b aria-hidden="true">⌄</b>
      </button>
      <nav aria-label="Collection navigation">
        <RouterLink :to="`/collections/${encodeURIComponent(collection)}`" exact-active-class="active">Overview</RouterLink>
        <RouterLink :to="`/collections/${encodeURIComponent(collection)}/files`">All files</RouterLink>
        <RouterLink :to="`/collections/${encodeURIComponent(collection)}/search`">Search</RouterLink>
        <RouterLink class="upload-link" :to="`/collections/${encodeURIComponent(collection)}/import`">Upload</RouterLink>
      </nav>
      <ThemeMenu />
    </header>
    <RouterView />
  </div>
</template>

<style scoped>
.app-shell { min-height: 100vh; }
.app-bar { display: grid; position: sticky; top: 0; z-index: 30; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: .5rem; min-height: 4rem; padding: .55rem .75rem; border-bottom: 1px solid var(--border); background: color-mix(in srgb, var(--surface) 94%, transparent); box-shadow: 0 1px 0 color-mix(in srgb, var(--primary) 8%, transparent); backdrop-filter: blur(.8rem); }
.brand { color: var(--primary); font: 800 1.15rem var(--display); letter-spacing: -.03em; text-decoration: none; }
.collection-button { display: flex; min-width: 0; min-height: 2.75rem; align-items: center; gap: .45rem; padding: .4rem .65rem; border: 1px solid var(--border); border-radius: var(--radius); color: var(--text); background: var(--surface-raised); cursor: pointer; }
.collection-button span { display: none; color: var(--text-muted); font-size: .75rem; }
.collection-button strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.collection-button b { color: var(--primary); }
nav { display: flex; position: fixed; inset: auto 0 0; z-index: 30; justify-content: space-around; padding: .35rem max(.5rem, env(safe-area-inset-right)) max(.35rem, env(safe-area-inset-bottom)) max(.5rem, env(safe-area-inset-left)); border-top: 1px solid var(--border); background: var(--surface); }
nav a { display: grid; min-height: 2.75rem; place-items: center; padding: .35rem .65rem; border-radius: var(--radius); color: var(--text-muted); font-size: .78rem; font-weight: 800; text-decoration: none; }
nav a.router-link-active, nav a.active { color: var(--primary); background: var(--primary-soft); }
nav .upload-link { color: var(--action); }
@media (min-width: 48rem) {
  .app-bar { grid-template-columns: auto auto 1fr auto; padding-inline: clamp(1rem, 3vw, 2.5rem); }
  .collection-button span { display: inline; }
  nav { display: flex; position: static; justify-content: flex-end; padding: 0; border: 0; background: transparent; }
  nav .upload-link { padding-inline: 1rem; color: white; background: var(--action); }
  nav .upload-link.router-link-active { color: white; background: var(--action-hover); }
}
</style>
