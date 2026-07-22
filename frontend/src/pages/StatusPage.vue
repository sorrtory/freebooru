<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'

import StatusDiagnostics from '../StatusDiagnostics.vue'
import ThemeMenu from '../components/ThemeMenu.vue'
import { useApplicationStatus } from '../useApplicationStatus'

const router = useRouter()
const { state, status, errorMessage, reload } = useApplicationStatus()

watch(status, (value) => {
  if (value?.ready && value.default_collection) {
    void router.replace(`/collections/${encodeURIComponent(value.default_collection)}`)
  }
})
onMounted(() => void reload())
</script>

<template>
  <main class="status-page">
    <div class="status-top"><span class="brand">FreeBooru</span><ThemeMenu /></div>
    <section v-if="state === 'loading'" class="status-content" role="status">
      <span class="loader" aria-hidden="true"></span><h1>Opening FreeBooru…</h1>
    </section>
    <section v-else-if="state === 'ready' && status && !status.ready" class="status-content">
      <p class="eyebrow">Configuration</p><h1>FreeBooru needs attention</h1>
      <StatusDiagnostics v-if="status.diagnostics.length" :diagnostics="status.diagnostics" />
      <button class="button button--primary" type="button" @click="reload">Check again</button>
    </section>
    <section v-else-if="state === 'error'" class="status-content" role="alert">
      <p class="eyebrow">Connection</p><h1>FreeBooru is unavailable</h1><p>{{ errorMessage }}</p>
      <button class="button button--primary" type="button" @click="reload">Retry</button>
    </section>
  </main>
</template>

<style scoped>
.status-page { min-height: 100vh; padding: 1rem; background: radial-gradient(circle at 78% 12%, var(--primary-soft), transparent 30rem), var(--canvas); }
.status-top { display: flex; align-items: center; justify-content: space-between; max-width: 72rem; margin: 0 auto; }
.brand { color: var(--primary); font: 800 1.2rem var(--display); }
.status-content { display: grid; width: min(100%, 42rem); min-height: calc(100vh - 8rem); align-content: center; gap: 1rem; margin: 0 auto; }
.status-content h1 { margin: 0; font: clamp(2.3rem, 8vw, 4.8rem)/.98 var(--display); letter-spacing: -.05em; }
.status-content > p:not(.eyebrow) { max-width: 55ch; color: var(--text-muted); line-height: 1.6; }
.eyebrow { margin: 0; color: var(--action); font-size: .78rem; font-weight: 800; letter-spacing: .1em; text-transform: uppercase; }
.button { justify-self: start; }
.loader { width: 2rem; height: 2rem; border: .2rem solid var(--primary-soft); border-top-color: var(--primary); border-radius: 50%; animation: spin .7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
