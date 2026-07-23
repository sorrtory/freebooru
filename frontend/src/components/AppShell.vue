<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import CollectionSwitcher from './CollectionSwitcher.vue'
import ThemeMenu from './ThemeMenu.vue'

const props = defineProps<{ collection: string }>()
const route = useRoute()
const router = useRouter()
const headerQuery = shallowRef('')

watch(() => route.fullPath, () => {
  headerQuery.value = route.path.endsWith('/files') ? String(route.query.q ?? '') : ''
}, { immediate: true })

function search() {
  const q = headerQuery.value.trim()
  void router.push({ path: `/collections/${encodeURIComponent(props.collection)}/files`, query: q ? { q } : {} })
}
</script>

<template>
  <div class="app-shell">
    <header class="app-bar">
      <RouterLink class="brand" to="/">FreeBooru</RouterLink>
      <CollectionSwitcher :collection="collection" />
      <RouterLink class="overview-link" :to="`/collections/${encodeURIComponent(collection)}`">Overview</RouterLink>
      <form class="header-search" role="search" @submit.prevent="search"><svg aria-hidden="true" viewBox="0 0 24 24"><circle cx="11" cy="11" r="6"/><path d="m16 16 4 4"/></svg><input v-model="headerQuery" type="search" aria-label="Search files" placeholder="Search tags"></form>
      <nav aria-label="Collection navigation">
        <RouterLink :to="`/collections/${encodeURIComponent(collection)}/storage`">Storage</RouterLink>
        <RouterLink :to="`/collections/${encodeURIComponent(collection)}/tags`">Tags</RouterLink>
        <RouterLink :to="`/collections/${encodeURIComponent(collection)}/files`">Files</RouterLink>
        <RouterLink class="upload-link" :to="`/collections/${encodeURIComponent(collection)}/import`"><svg aria-hidden="true" viewBox="0 0 24 24"><path d="M12 16V4m0 0L7.5 8.5M12 4l4.5 4.5M4 14v5h16v-5"/></svg><span>Upload</span></RouterLink>
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
.app-bar > .overview-link { display: none; color: var(--text-muted); font-size: .8rem; font-weight: 800; text-decoration: none; }
.app-bar > .overview-link.router-link-exact-active { color: var(--primary); }
.header-search { display: flex; grid-column: 1 / -1; align-items: center; min-height: 2.75rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--background); }
.header-search:focus-within { border-color: var(--primary); box-shadow: 0 0 0 3px var(--primary-soft); }
.header-search svg { width: 1.05rem; margin-left: .7rem; fill: none; stroke: var(--text-muted); stroke-width: 2; }
.header-search input { width: 100%; min-width: 0; padding: .55rem .7rem; border: 0; outline: 0; color: var(--text); background: transparent; font: inherit; }
nav { display: flex; position: fixed; inset: auto 0 0; z-index: 30; justify-content: space-around; padding: .35rem max(.5rem, env(safe-area-inset-right)) max(.35rem, env(safe-area-inset-bottom)) max(.5rem, env(safe-area-inset-left)); border-top: 1px solid var(--border); background: var(--surface); }
nav a { display: flex; min-width: 0; min-height: 2.75rem; flex: 1; align-items: center; justify-content: center; gap: .25rem; padding: .3rem .2rem; border-radius: var(--radius); color: var(--text-muted); font-size: clamp(.62rem, 2.5vw, .78rem); font-weight: 800; text-decoration: none; }
nav a:not(.overview-link).router-link-active, nav a.router-link-exact-active { color: var(--primary); background: var(--primary-soft); }
nav svg { width: 1.05rem; fill: none; stroke: currentColor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 1.8; }
nav .upload-link { color: var(--primary); }
@media (min-width: 48rem) {
  .app-bar { grid-template-columns: auto auto auto minmax(10rem, 1fr) auto auto; padding-inline: clamp(1rem, 3vw, 2.5rem); }
  .app-bar > .overview-link { display: flex; min-height: 2.75rem; align-items: center; padding-inline: .55rem; }
  .header-search { grid-column: auto; max-width: 30rem; margin-inline: auto; }
  nav { display: flex; position: static; justify-content: flex-end; padding: 0; border: 0; background: transparent; }
  nav a { flex: 0 1 auto; padding: .35rem .65rem; font-size: .78rem; }
  nav .upload-link { padding-inline: 1rem; color: white; background: var(--primary); }
  nav .upload-link.router-link-active { color: white; background: var(--primary-hover); }
}
</style>
