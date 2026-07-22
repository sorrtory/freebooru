<script setup lang="ts">
import { onMounted, shallowRef } from 'vue'

import { getHello, type HelloResponse } from './api'

type LoadState = 'loading' | 'ready' | 'error'

const state = shallowRef<LoadState>('loading')
const hello = shallowRef<HelloResponse>()
const errorMessage = shallowRef('')

async function loadHello() {
  state.value = 'loading'
  errorMessage.value = ''

  try {
    hello.value = await getHello()
    state.value = 'ready'
  } catch (error) {
    hello.value = undefined
    errorMessage.value = error instanceof Error ? error.message : 'Unable to reach FreeBooru'
    state.value = 'error'
  }
}

onMounted(() => void loadHello())
</script>

<template>
  <main class="shell">
    <section class="hello-card" aria-labelledby="app-title">
      <div class="eyebrow">
        <span class="status-dot" :class="`status-dot--${state}`" aria-hidden="true"></span>
        <span>Personal archive / system check</span>
      </div>

      <div class="title-block">
        <p class="kicker">FreeBooru</p>
        <h1 id="app-title">
          <template v-if="state === 'loading'">Connecting…</template>
          <template v-else-if="state === 'ready'">{{ hello?.message }}</template>
          <template v-else>Connection interrupted</template>
        </h1>
      </div>

      <div v-if="state === 'loading'" class="detail" role="status">
        <span class="loader" aria-hidden="true"></span>
        Contacting the local API
      </div>

      <div v-else-if="state === 'ready'" class="detail detail--ready">
        <span>Transport</span>
        <strong>{{ hello?.mode === 'desktop' ? 'Wails desktop' : 'Web server' }}</strong>
      </div>

      <div v-else class="error-panel" role="alert">
        <p>{{ errorMessage }}</p>
        <button type="button" @click="loadHello">Try again</button>
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

.hello-card {
  position: relative;
  width: min(100%, 42rem);
  overflow: hidden;
  padding: clamp(1.5rem, 6vw, 4rem);
  border: 1px solid var(--line);
  border-radius: 1.5rem;
  background: linear-gradient(145deg, rgb(31 35 29 / 96%), rgb(18 20 17 / 98%));
  box-shadow: 0 1.75rem 5rem rgb(0 0 0 / 38%);
}

.hello-card::before {
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
  margin: clamp(3.5rem, 12vw, 6.5rem) 0 clamp(2.5rem, 8vw, 4.5rem);
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
  max-width: 12ch;
  margin: 0;
  font-size: clamp(2.5rem, 10vw, 5.4rem);
  font-weight: 540;
  letter-spacing: -0.065em;
  line-height: 0.94;
}

.detail {
  display: flex;
  min-height: 3.25rem;
  align-items: center;
  gap: 0.8rem;
  padding: 0.85rem 1rem;
  border: 1px solid var(--line);
  border-radius: 0.8rem;
  color: var(--muted);
  font-family: var(--mono);
  font-size: 0.86rem;
}

.detail--ready {
  justify-content: space-between;
}

.detail strong {
  color: var(--text);
  font-weight: 500;
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
