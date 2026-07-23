<script setup lang="ts">
import { computed, onMounted, ref, shallowRef } from 'vue'
import { onBeforeRouteLeave, useRoute } from 'vue-router'

import { getSettings, updateSettings, type ApplicationSettings } from '../api'
import ValueCombobox from '../components/ValueCombobox.vue'
import { useTheme, type ThemePreference } from '../useTheme'

const { preference, select } = useTheme()
const route = useRoute()
const choices: { value: ThemePreference; label: string; description: string }[] = [
  { value: 'system', label: 'System', description: 'Follow this device' },
  { value: 'light', label: 'Light', description: 'Bright archive surfaces' },
  { value: 'dark', label: 'Dark', description: 'Dim archive surfaces' },
]
const settings = ref<ApplicationSettings>()
const saved = ref<ApplicationSettings>()
const loading = shallowRef(true)
const saving = shallowRef(false)
const error = shallowRef('')
const notice = shallowRef('')
const dirty = computed(() => Boolean(settings.value && saved.value && JSON.stringify(settings.value) !== JSON.stringify(saved.value)))
const storageCollection = computed(() => typeof route.query.collection === 'string' ? route.query.collection : settings.value?.default_collection ?? '')

onMounted(load)
onBeforeRouteLeave(() => !dirty.value || window.confirm('Discard unsaved settings?'))

async function load() {
  loading.value = true; error.value = ''
  try { const value = await getSettings(); settings.value = { ...value }; saved.value = { ...value } }
  catch (cause) { error.value = cause instanceof Error ? cause.message : 'Settings are unavailable.' }
  finally { loading.value = false }
}

async function save() {
  if (!settings.value || !dirty.value) return
  saving.value = true; error.value = ''; notice.value = ''
  try {
    const value = await updateSettings(settings.value)
    settings.value = { ...value }; saved.value = { ...value }
    notice.value = value.restart_required ? 'Saved. Restart FreeBooru to use the new HTTP port.' : 'Settings saved.'
  } catch (cause) { error.value = cause instanceof Error ? cause.message : 'Settings could not be saved.' }
  finally { saving.value = false }
}
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
    <section aria-labelledby="application-heading">
      <div><p>FreeBooru</p><h2 id="application-heading">Application</h2></div>
      <p v-if="loading" role="status">Loading settings…</p>
      <form v-else-if="settings" @submit.prevent="save">
        <fieldset><legend>Language</legend><label class="inline-choice"><input v-model="settings.language" type="radio" value="en"> English</label><label class="inline-choice"><input v-model="settings.language" type="radio" value="ru"> Russian</label></fieldset>
        <label class="form-field"><span>Default collection</span><ValueCombobox v-model="settings.default_collection" :options="settings.collections" /></label>
        <label class="form-field"><span>Default storage</span><ValueCombobox v-model="settings.default_storage_name" :options="settings.storages" /></label>
        <label class="form-field"><span>HTTP port <small>Requires restart</small></span><input v-model.number="settings.http_port" type="number" min="1" max="65535" step="1"></label>
        <label class="check-field"><input v-model="settings.remove_on_upload" type="checkbox"><span><strong>Remove source after upload</strong><small>Only after stored content and database commit are verified</small></span></label>
        <div class="save-row"><button class="button button--primary" type="submit" :disabled="saving || !dirty">{{ saving ? 'Saving…' : 'Save' }}</button><span v-if="dirty">Unsaved changes</span></div>
      </form>
    </section>
    <section aria-labelledby="storage-heading">
      <div><p>Collection</p><h2 id="storage-heading">Storage</h2></div>
      <div class="settings-action"><p>Inspect locations available to {{ storageCollection || 'the selected collection' }}.</p><RouterLink v-if="storageCollection" class="button" :to="`/collections/${encodeURIComponent(storageCollection)}/storage`">Manage storage</RouterLink></div>
    </section>
    <p v-if="error" class="settings-error" role="alert">{{ error }} <button class="button" type="button" @click="load">Reload</button></p><p v-if="notice" class="settings-notice" role="status">{{ notice }}</p>
  </main>
</template>

<style scoped>
.settings-page { display: grid; max-width: 58rem; gap: 1rem; }
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
input { accent-color: var(--primary); }form { display: grid; gap: 1rem; }form fieldset { display: flex; flex-wrap: wrap; gap: .5rem; }.inline-choice { min-height: 2.75rem; }.form-field { display: grid; gap: .4rem; padding: 0; border: 0; }.form-field > span { font-weight: 800; }.form-field small { color: var(--text-muted); font-weight: 400; }.form-field > input { width: 100%; padding: .65rem .75rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--canvas); }.check-field { align-items: flex-start; }.check-field span { display: grid; }.check-field small { color: var(--text-muted); }.save-row { display: flex; align-items: center; gap: .75rem; }.save-row span { color: var(--text-muted); }.settings-error { color: var(--danger); }.settings-notice { color: var(--success); font-weight: 800; }
.settings-action { display: grid; justify-items: start; gap: .75rem; }.settings-action p { color: var(--text-muted); font-size: .9rem; font-weight: 400; letter-spacing: 0; text-transform: none; }
@media (min-width: 48rem) { section { grid-template-columns: 12rem 1fr; } }
</style>
