<script setup lang="ts">
import { useTheme, type ThemePreference } from '../useTheme'

const { preference, select } = useTheme()
const choices: { value: ThemePreference; label: string; description: string }[] = [
  { value: 'system', label: 'System', description: 'Follow this device' },
  { value: 'light', label: 'Light', description: 'Bright archive surfaces' },
  { value: 'dark', label: 'Dark', description: 'Dim archive surfaces' },
]
</script>

<template>
  <main class="settings-page page">
    <header class="settings-header"><RouterLink to="/" aria-label="Return to FreeBooru">FreeBooru</RouterLink><h1>Settings</h1></header>
    <section aria-labelledby="appearance-heading">
      <div><p>Interface</p><h2 id="appearance-heading">Appearance</h2></div>
      <fieldset>
        <legend class="visually-hidden">Theme</legend>
        <label v-for="choice in choices" :key="choice.value" :class="{ selected: preference === choice.value }">
          <input type="radio" name="theme" :value="choice.value" :checked="preference === choice.value" @change="select(choice.value)">
          <span><strong>{{ choice.label }}</strong><small>{{ choice.description }}</small></span>
        </label>
      </fieldset>
    </section>
  </main>
</template>

<style scoped>
.settings-page { max-width: 58rem; }
.settings-header { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; margin-bottom: 2rem; }
.settings-header a { color: var(--primary); font: 800 1rem var(--display); text-decoration: none; }
.settings-header h1 { margin: 0; font: clamp(2.25rem, 6vw, 4rem)/1 var(--display); }
section { display: grid; gap: 1.25rem; padding: clamp(1rem, 3vw, 1.5rem); border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface); }
section p { margin: 0 0 .2rem; color: var(--text-muted); font-size: .75rem; font-weight: 800; letter-spacing: .08em; text-transform: uppercase; }
section h2 { margin: 0; }
fieldset { display: grid; gap: .5rem; margin: 0; padding: 0; border: 0; }
label { display: flex; min-height: 3.5rem; align-items: center; gap: .75rem; padding: .65rem .8rem; border: 1px solid var(--border); border-radius: var(--radius); cursor: pointer; }
label.selected { border-color: var(--primary); background: var(--primary-soft); }
label span { display: grid; gap: .15rem; }
label small { color: var(--text-muted); }
input { accent-color: var(--primary); }
@media (min-width: 48rem) { section { grid-template-columns: 12rem 1fr; } }
</style>
