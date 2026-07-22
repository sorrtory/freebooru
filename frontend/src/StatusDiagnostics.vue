<script setup lang="ts">
import type { Diagnostic } from './api'

defineProps<{
  diagnostics: readonly Diagnostic[]
}>()

function source(diagnostic: Diagnostic): string {
  return [
    diagnostic.file,
    diagnostic.document > 0 ? `document ${diagnostic.document}` : '',
    diagnostic.field,
  ]
    .filter(Boolean)
    .join(' · ')
}
</script>

<template>
  <ul class="diagnostics" aria-label="Configuration diagnostics">
    <li
      v-for="diagnostic in diagnostics"
      :key="`${diagnostic.code}:${diagnostic.file}:${diagnostic.document}:${diagnostic.field}:${diagnostic.message}`"
      class="diagnostic"
      :class="`diagnostic--${diagnostic.severity}`"
    >
      <div class="diagnostic__heading">
        <span>{{ diagnostic.severity }}</span>
        <code>{{ diagnostic.code }}</code>
      </div>
      <p>{{ diagnostic.message }}</p>
      <small v-if="source(diagnostic)">{{ source(diagnostic) }}</small>
    </li>
  </ul>
</template>

<style scoped>
.diagnostics {
  display: grid;
  gap: 0.75rem;
  max-height: 16rem;
  margin: 1rem 0 0;
  padding: 0;
  overflow-y: auto;
  list-style: none;
}

.diagnostic {
  padding: 0.9rem 1rem;
  border: 1px solid var(--line);
  border-left: 0.2rem solid var(--muted);
  border-radius: 0.65rem;
  background: rgb(255 255 255 / 2%);
}

.diagnostic--error {
  border-left-color: var(--error);
}

.diagnostic--warning {
  border-left-color: var(--warning);
}

.diagnostic__heading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.6rem;
  font: 0.72rem var(--mono);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.diagnostic__heading span {
  color: var(--muted);
}

.diagnostic code {
  color: var(--text);
  font: inherit;
  text-transform: none;
}

.diagnostic p {
  margin: 0.55rem 0 0;
  line-height: 1.45;
}

.diagnostic small {
  display: block;
  margin-top: 0.45rem;
  color: var(--muted);
  font: 0.72rem/1.4 var(--mono);
  overflow-wrap: anywhere;
}
</style>
