<script setup lang="ts">
import { computed, onMounted } from 'vue'

import StatusDiagnostics from './StatusDiagnostics.vue'
import { useApplicationStatus } from './useApplicationStatus'

const { state, status, errorMessage, reload } = useApplicationStatus()

const presentation = computed(() => {
  if (state.value === 'loading') {
    return { title: 'Checking configuration…', tone: 'loading' }
  }
  if (state.value === 'error') {
    return { title: 'Connection interrupted', tone: 'error' }
  }
  if (status.value?.ready) {
    return { title: 'FreeBooru is ready', tone: 'ready' }
  }
  return { title: 'Configuration needs attention', tone: 'error' }
})

onMounted(() => void reload())
</script>

<template>
  <main class="shell">
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
          <div>
            <dt>Runtime</dt>
            <dd>{{ status.mode === 'desktop' ? 'Wails desktop' : 'Web server' }}</dd>
          </div>
          <div>
            <dt>Collection</dt>
            <dd>{{ status.default_collection || 'Unavailable' }}</dd>
          </div>
        </dl>

        <StatusDiagnostics
          v-if="status.diagnostics.length > 0"
          :diagnostics="status.diagnostics"
        />

        <button v-if="!status.ready" type="button" @click="reload">Retry</button>
      </template>

      <div v-else class="error-panel" role="alert">
        <p>{{ errorMessage }}</p>
        <button type="button" @click="reload">Retry</button>
      </div>

      <footer>
        <span>One interface</span>
        <span aria-hidden="true">·</span>
        <span>Two runtimes</span>
      </footer>
    </section>
  </main>
</template>

<style scoped>
.shell {
  display: grid;
  min-height: 100%;
  place-items: center;
  padding: 1.25rem;
}

.status-card {
  position: relative;
  width: min(100%, 46rem);
  overflow: hidden;
  padding: clamp(1.5rem, 6vw, 4rem);
  border: 1px solid var(--line);
  border-radius: 1.5rem;
  background: linear-gradient(145deg, rgb(31 35 29 / 96%), rgb(18 20 17 / 98%));
  box-shadow: 0 1.75rem 5rem rgb(0 0 0 / 38%);
}

.status-card::before {
  position: absolute;
  inset: 0 0 auto;
  height: 0.2rem;
  background: linear-gradient(90deg, var(--accent), transparent 72%);
  content: '';
}

.eyebrow,
footer {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  color: var(--muted);
  font-family: var(--mono);
  font-size: 0.75rem;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.status-dot {
  width: 0.55rem;
  height: 0.55rem;
  flex: 0 0 auto;
  border-radius: 50%;
  background: var(--muted);
}

.status-dot--loading {
  animation: pulse 1.2s ease-in-out infinite;
}

.status-dot--ready {
  background: var(--accent);
  box-shadow: 0 0 1rem rgb(181 240 99 / 60%);
}

.status-dot--error {
  background: var(--error);
}

.title-block {
  margin: clamp(3rem, 10vw, 5.5rem) 0 clamp(2rem, 7vw, 3.5rem);
}

.kicker {
  margin: 0 0 0.8rem;
  color: var(--accent);
  font-family: var(--mono);
  font-size: 0.82rem;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

h1 {
  max-width: 15ch;
  margin: 0;
  font-size: clamp(2.35rem, 9vw, 4.8rem);
  font-weight: 540;
  letter-spacing: -0.06em;
  line-height: 0.96;
}

.detail,
.runtime-details {
  min-height: 3.25rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--line);
  border-radius: 0.8rem;
  font-family: var(--mono);
  font-size: 0.86rem;
}

.detail {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  color: var(--muted);
}

.runtime-details {
  display: grid;
  gap: 1rem;
  margin: 0;
}

.runtime-details div {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
}

.runtime-details dt {
  color: var(--muted);
}

.runtime-details dd {
  margin: 0;
  text-align: right;
}

.loader {
  width: 1rem;
  height: 1rem;
  border: 2px solid var(--line);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.75s linear infinite;
}

.error-panel {
  padding: 1rem;
  border: 1px solid rgb(255 125 105 / 35%);
  border-radius: 0.8rem;
  background: rgb(255 125 105 / 6%);
}

.error-panel p {
  margin: 0 0 1rem;
  color: #ffc1b8;
  line-height: 1.5;
}

button {
  min-width: 7rem;
  min-height: 2.75rem;
  margin-top: 1rem;
  padding: 0.65rem 1rem;
  border: 0;
  border-radius: 0.6rem;
  background: var(--accent);
  color: #15190f;
  cursor: pointer;
  font: 700 0.85rem var(--mono);
}

button:hover {
  background: #c9ff80;
}

button:focus-visible {
  outline: 3px solid rgb(181 240 99 / 38%);
  outline-offset: 3px;
}

footer {
  margin-top: 1.5rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse {
  50% {
    opacity: 0.35;
  }
}

@media (min-width: 36rem) {
  .runtime-details {
    grid-template-columns: 1fr 1fr;
  }
}

@media (min-width: 48rem) {
  .shell {
    padding: 3rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .status-dot--loading,
  .loader {
    animation: none;
  }
}
</style>
