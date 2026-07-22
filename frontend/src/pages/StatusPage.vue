<script setup lang="ts">
import { computed, onMounted } from 'vue'

import StatusDiagnostics from '../StatusDiagnostics.vue'
import { useApplicationStatus } from '../useApplicationStatus'

const { state, status, errorMessage, reload } = useApplicationStatus()
const presentation = computed(() => {
  if (state.value === 'loading') return { title: 'Checking configuration…', tone: 'loading' }
  if (state.value === 'error') return { title: 'Connection interrupted', tone: 'error' }
  if (status.value?.ready) return { title: 'FreeBooru is ready', tone: 'ready' }
  return { title: 'Configuration needs attention', tone: 'error' }
})

onMounted(() => void reload())
</script>

<template>
  <main class="status-shell">
    <section class="status-card" aria-labelledby="app-title">
      <div class="eyebrow">
        <span class="status-dot" :class="`status-dot--${presentation.tone}`" aria-hidden="true"></span>
        <span>Personal archive / system check</span>
      </div>
      <div class="title-block">
        <p class="kicker">FreeBooru</p>
        <h1 id="app-title">{{ presentation.title }}</h1>
      </div>
      <div v-if="state === 'loading'" class="detail" role="status">
        <span class="loader" aria-hidden="true"></span>
        Loading the application configuration
      </div>
      <template v-else-if="state === 'ready' && status">
        <dl class="runtime-details">
          <div><dt>Runtime</dt><dd>{{ status.mode === 'desktop' ? 'Wails desktop' : 'Web server' }}</dd></div>
          <div><dt>Collection</dt><dd>{{ status.default_collection || 'Unavailable' }}</dd></div>
        </dl>
        <StatusDiagnostics v-if="status.diagnostics.length" :diagnostics="status.diagnostics" />
        <p v-if="status.diagnostics.length" class="diagnostic-count">
          {{ status.diagnostics.length }} {{ status.diagnostics.length === 1 ? 'diagnostic' : 'diagnostics' }}
        </p>
        <RouterLink v-if="status.ready" class="primary-link" :to="`/collections/${encodeURIComponent(status.default_collection)}/import`">Import a file</RouterLink>
        <button v-else type="button" @click="reload">Retry</button>
      </template>
      <div v-else class="error-panel" role="alert">
        <p>{{ errorMessage }}</p>
        <button type="button" @click="reload">Retry</button>
      </div>
      <footer><span>One interface</span><span aria-hidden="true">·</span><span>Two runtimes</span></footer>
    </section>
  </main>
</template>

<style scoped>
.status-shell { display: grid; min-height: 100%; place-items: center; padding: 1.25rem; }
.status-card { position: relative; width: min(100%, 46rem); overflow: hidden; padding: clamp(1.5rem, 6vw, 4rem); border: 1px solid var(--line); border-radius: 1.5rem; background: linear-gradient(145deg, rgb(31 35 29 / 96%), rgb(18 20 17 / 98%)); box-shadow: 0 1.75rem 5rem rgb(0 0 0 / 38%); }
.status-card::before { position: absolute; inset: 0 0 auto; height: .2rem; background: linear-gradient(90deg, var(--accent), transparent 72%); content: ''; }
.eyebrow, footer { display: flex; align-items: center; gap: .6rem; color: var(--muted); font: .75rem var(--mono); letter-spacing: .08em; text-transform: uppercase; }
.status-dot { width: .55rem; height: .55rem; flex: 0 0 auto; border-radius: 50%; background: var(--muted); }
.status-dot--loading { animation: pulse 1.2s ease-in-out infinite; }
.status-dot--ready { background: var(--accent); box-shadow: 0 0 1rem rgb(181 240 99 / 60%); }
.status-dot--error { background: var(--error); }
.title-block { margin: clamp(3rem, 10vw, 5.5rem) 0 clamp(2rem, 7vw, 3.5rem); }
.kicker { margin: 0 0 .8rem; color: var(--accent); font: .82rem var(--mono); letter-spacing: .14em; text-transform: uppercase; }
h1 { max-width: 15ch; margin: 0; font: 540 clamp(2.35rem, 9vw, 4.8rem)/.96 var(--display); letter-spacing: -.06em; }
.detail, .runtime-details { min-height: 3.25rem; padding: .85rem 1rem; border: 1px solid var(--line); border-radius: .8rem; font: .86rem var(--mono); }
.detail { display: flex; align-items: center; gap: .8rem; color: var(--muted); }
.runtime-details { display: grid; gap: 1rem; margin: 0; }
.runtime-details div { display: flex; justify-content: space-between; gap: 1rem; }
.runtime-details dt { color: var(--muted); }
.runtime-details dd { margin: 0; text-align: right; }
.loader { width: 1rem; height: 1rem; border: 2px solid var(--line); border-top-color: var(--accent); border-radius: 50%; animation: spin .75s linear infinite; }
.error-panel { padding: 1rem; border: 1px solid rgb(255 125 105 / 35%); border-radius: .8rem; background: rgb(255 125 105 / 6%); }
.error-panel p { margin: 0 0 1rem; color: #ffc1b8; line-height: 1.5; }
.diagnostic-count { margin: .6rem 0 0; color: var(--muted); font: .72rem var(--mono); text-align: right; }
.primary-link { display: inline-flex; align-items: center; min-height: 2.75rem; margin-top: 1rem; padding: .65rem 1rem; border-radius: .6rem; background: var(--accent); color: #15190f; font: 700 .85rem var(--mono); text-decoration: none; }
button { min-width: 7rem; margin-top: 1rem; }
footer { margin-top: 1.5rem; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
