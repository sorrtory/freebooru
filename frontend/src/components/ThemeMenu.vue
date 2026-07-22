<script setup lang="ts">
import { useTheme } from '../useTheme'
import type { ThemePreference } from '../useTheme'

const { preference, resolved, select } = useTheme()
const choices: { value: ThemePreference; label: string }[] = [
  { value: 'system', label: 'System' },
  { value: 'light', label: 'Light' },
  { value: 'dark', label: 'Dark' },
]
</script>

<template>
  <details class="theme-menu">
    <summary :aria-label="`Theme: ${preference}`" title="Theme">
      <span aria-hidden="true">{{ resolved === 'dark' ? '◐' : '◑' }}</span>
      <span class="wide-label">Theme</span>
    </summary>
    <fieldset>
      <legend>Appearance</legend>
      <label v-for="choice in choices" :key="choice.value">
        <input type="radio" name="theme" :value="choice.value" :checked="preference === choice.value" @change="select(choice.value)">
        {{ choice.label }}
      </label>
    </fieldset>
  </details>
</template>

<style scoped>
.theme-menu { position: relative; }
summary { display: flex; min-width: 2.75rem; min-height: 2.75rem; align-items: center; justify-content: center; gap: .45rem; padding: 0 .75rem; border: 1px solid var(--border); border-radius: var(--radius); color: var(--text); background: var(--surface); cursor: pointer; font-weight: 700; list-style: none; }
summary::-webkit-details-marker { display: none; }
fieldset { display: grid; position: absolute; top: calc(100% + .5rem); right: 0; z-index: 20; width: 10rem; gap: .15rem; margin: 0; padding: .5rem; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); box-shadow: var(--shadow); }
legend { padding: .35rem .45rem; color: var(--text-muted); font-size: .75rem; font-weight: 700; text-transform: uppercase; }
label { display: flex; min-height: 2.75rem; align-items: center; gap: .65rem; padding: .5rem; border-radius: var(--radius); cursor: pointer; }
label:hover { background: var(--primary-soft); }
input { accent-color: var(--primary); }
@media (max-width: 47.999rem) { .wide-label { position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); } }
</style>
