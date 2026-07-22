<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, shallowRef } from 'vue'

import { APIError, importFile } from '../api'
import type { ImportResult, Relationship, TagValue } from '../api'
import AssignedTags from '../components/AssignedTags.vue'
import ImportGuidance from '../components/ImportGuidance.vue'
import TagCatalog from '../components/TagCatalog.vue'
import { useImportDraft } from '../useImportDraft'

const props = defineProps<{ collection: string }>()
const { schema, draft, loading, evaluating, errorMessage, load, apply, remove } = useImportDraft(props.collection)
const file = shallowRef<File>()
const previewURL = shallowRef('')
const activePanel = shallowRef<'assigned' | 'suggested' | 'add'>('assigned')
const selectedTag = shallowRef('')
const uploading = shallowRef(false)
const uploadResult = shallowRef<ImportResult>()
const uploadError = shallowRef('')

const assignedNames = computed(() => new Set(draft.value?.assignments.map((item) => item.name) ?? []))
const representedNames = computed(() => new Set([
  ...assignedNames.value,
  ...(draft.value?.missing_required.map((item) => item.name) ?? []),
]))
const conflictNames = computed(() => new Set(draft.value?.active_conflicts.flatMap((edge) => [edge.source_tag, edge.target_tag]) ?? []))
const issueCount = computed(() => (draft.value?.missing_required.length ?? 0) + (draft.value?.missing_demands.length ?? 0) + (draft.value?.active_conflicts.length ?? 0))
const imagePreview = computed(() => file.value?.type.startsWith('image/') ?? false)

function chooseFile(event: Event) {
  const next = (event.target as HTMLInputElement).files?.[0]
  if (!next) return
  if (previewURL.value) URL.revokeObjectURL(previewURL.value)
  file.value = next
  previewURL.value = URL.createObjectURL(next)
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 ** 2).toFixed(1)} MB`
}

function chooseRelationship(edge: Relationship) {
  const exact = edge.target.is
  if (exact !== undefined) {
    void apply(edge.target_tag, exact)
    return
  }
  if (edge.target.has.length) {
    void apply(edge.target_tag, edge.target.has)
    return
  }
  const field = schema.value?.fields.find((item) => item.name === edge.target_tag)
  if (edge.target.presence && field?.type === 'bool') {
    void apply(edge.target_tag, true)
    return
  }
  if (draft.value?.missing_required.some((item) => item.name === edge.target_tag)) {
    activePanel.value = 'assigned'
    return
  }
  selectedTag.value = edge.target_tag
  activePanel.value = 'add'
}

function applyTag(name: string, value: TagValue) {
  void apply(name, value)
}

async function submit() {
  if (!file.value || !draft.value?.complete || uploading.value) return
  uploading.value = true
  uploadError.value = ''
  try {
    uploadResult.value = await importFile(props.collection, file.value, Object.fromEntries(draft.value.assignments.map((item) => [item.name, item.value])))
  } catch (error) {
    uploadError.value = error instanceof APIError && error.code === 'import.duplicate'
      ? `This content already exists: ${error.message}`
      : error instanceof Error ? error.message : 'Import failed'
  } finally {
    uploading.value = false
  }
}

onMounted(() => void load())
onBeforeUnmount(() => { if (previewURL.value) URL.revokeObjectURL(previewURL.value) })
</script>

<template>
  <main class="import-shell">
    <header class="import-header">
      <div class="file-picker">
        <div v-if="file" class="file-summary">
          <img v-if="imagePreview" :src="previewURL" alt="Selected file preview">
          <span v-else class="file-glyph" aria-hidden="true">FILE</span>
          <div><strong>{{ file.name }}</strong><small>{{ formatSize(file.size) }} · {{ file.type || 'unknown type' }}</small></div>
        </div>
        <div v-else><p class="kicker">New archive object</p><h1>Prepare one file</h1></div>
        <label class="button file-button" for="import-file">{{ file ? 'Replace file' : 'Choose file' }}</label>
        <input id="import-file" class="visually-hidden" type="file" @change="chooseFile">
      </div>
    </header>

    <div v-if="loading" class="workspace-state" role="status">Reading the {{ collection }} tag schema…</div>
    <div v-else-if="errorMessage && !draft" class="workspace-state workspace-state--error" role="alert">
      <p>{{ errorMessage }}</p><button class="button" type="button" @click="load">Retry</button>
    </div>
    <template v-else-if="schema && draft">
      <nav class="panel-tabs" aria-label="Import workspace panels">
        <button v-for="panel in (['assigned', 'suggested', 'add'] as const)" :key="panel" type="button" :aria-current="activePanel === panel ? 'page' : undefined" @click="activePanel = panel">
          {{ panel === 'add' ? 'Add' : panel }}
          <span v-if="panel === 'assigned'">{{ draft.assignments.length + draft.missing_required.length }}</span>
          <span v-else-if="panel === 'suggested'">{{ draft.missing_demands.length + draft.suggestions.length }}</span>
          <span v-else>{{ schema.fields.length - representedNames.size }}</span>
        </button>
      </nav>
      <div class="workspace-grid" :aria-busy="evaluating">
        <div :class="{ 'mobile-hidden': activePanel !== 'assigned' }">
          <AssignedTags :missing="draft.missing_required" :assigned="draft.assignments" :fields="schema.fields" :conflicts="conflictNames" @apply="applyTag" @remove="remove" />
        </div>
        <div :class="{ 'mobile-hidden': activePanel !== 'suggested' }">
          <ImportGuidance :demands="draft.missing_demands" :suggestions="draft.suggestions" @choose="chooseRelationship" />
        </div>
        <div :class="{ 'mobile-hidden': activePanel !== 'add' }">
          <TagCatalog :fields="schema.fields" :assigned="representedNames" :selected="selectedTag" @apply="applyTag" />
        </div>
      </div>
      <p v-if="errorMessage || uploadError" class="inline-error" role="alert">{{ errorMessage || uploadError }}</p>
      <footer class="action-bar">
        <div>
          <strong v-if="uploadResult">Imported {{ uploadResult.sha256.slice(0, 12) }}</strong>
          <strong v-else-if="issueCount">{{ issueCount }} blocking {{ issueCount === 1 ? 'issue' : 'issues' }}</strong>
          <strong v-else>Draft valid</strong>
          <span v-if="draft.active_conflicts.length">{{ draft.active_conflicts.length }} conflict{{ draft.active_conflicts.length === 1 ? '' : 's' }}</span>
          <span v-else-if="uploading">Streaming file to FreeBooru…</span>
          <span v-else-if="evaluating">Checking relationships…</span>
          <span v-else>{{ file ? file.name : 'Choose a file to continue' }}</span>
        </div>
        <button class="button button--primary" type="button" :disabled="!file || !draft.complete || evaluating || uploading || Boolean(uploadResult)" @click="submit">{{ uploading ? 'Importing…' : uploadResult ? 'Imported' : 'Import file' }}</button>
      </footer>
    </template>
  </main>
</template>

<style scoped>
.import-shell { min-height: 100vh; padding-bottom: 6.5rem; background: linear-gradient(90deg, color-mix(in srgb, var(--border) 18%, transparent) 1px, transparent 1px) 0 0 / 3rem 3rem; }
.import-header { border-bottom: 1px solid var(--line-strong); background: color-mix(in srgb, var(--surface) 94%, transparent); }
.file-picker { display: flex; min-height: 8rem; align-items: center; justify-content: space-between; gap: 1rem; padding: 1rem clamp(1rem, 3vw, 2rem); }
.file-picker h1 { margin: .25rem 0 0; font: clamp(2rem, 5vw, 3.7rem)/1 var(--display); letter-spacing: -.045em; }
.file-summary { display: flex; min-width: 0; align-items: center; gap: 1rem; }
.file-summary img, .file-glyph { width: 5rem; height: 5rem; flex: 0 0 auto; border: 1px solid var(--line-strong); border-radius: .3rem; object-fit: cover; }
.file-glyph { display: grid; place-items: center; color: var(--muted); font: .7rem var(--mono); }
.file-summary div { display: grid; min-width: 0; gap: .35rem; }
.file-summary strong { overflow: hidden; font-size: 1.05rem; text-overflow: ellipsis; white-space: nowrap; }
.file-summary small { color: var(--muted); font-size: .78rem; }
.file-button { flex: 0 0 auto; cursor: pointer; }
.panel-tabs { display: grid; grid-template-columns: repeat(3, 1fr); position: sticky; top: 0; z-index: 4; border-bottom: 1px solid var(--line-strong); background: var(--surface); }
.panel-tabs button { min-height: 3rem; border: 0; border-right: 1px solid var(--line); color: var(--muted); background: transparent; font: 700 .7rem var(--mono); text-transform: uppercase; }
.panel-tabs button[aria-current='page'] { color: var(--accent); box-shadow: inset 0 -2px var(--accent); }
.panel-tabs span { display: inline-grid; min-width: 1.3rem; min-height: 1.3rem; margin-left: .25rem; place-items: center; border-radius: 50%; background: var(--surface-raised); }
.workspace-grid { max-width: 112rem; margin: 0 auto; padding: 1rem; }
.workspace-grid[aria-busy='true'] { opacity: .72; transition: opacity .12s ease; }
.workspace-state { display: grid; min-height: 50vh; place-items: center; padding: 2rem; color: var(--muted); font: .8rem var(--mono); }
.workspace-state--error { color: var(--danger); }
.mobile-hidden { display: none; }
.inline-error { position: fixed; right: 1rem; bottom: 6rem; z-index: 8; max-width: 30rem; padding: .75rem 1rem; border: 1px solid var(--error); background: color-mix(in srgb, var(--danger) 10%, var(--surface)); color: var(--danger); }
.action-bar { display: flex; position: fixed; inset: auto 0 0; z-index: 6; align-items: center; justify-content: space-between; gap: 1rem; min-height: 5rem; padding: .75rem max(1rem, env(safe-area-inset-right)) max(.75rem, env(safe-area-inset-bottom)) max(1rem, env(safe-area-inset-left)); border-top: 1px solid var(--line-strong); background: color-mix(in srgb, var(--surface) 96%, transparent); backdrop-filter: blur(1rem); }
.action-bar div { display: grid; gap: .2rem; }
.action-bar strong { color: var(--primary); font-size: .82rem; }
.action-bar span { color: var(--muted); font-size: .8rem; }
.action-bar .button { min-width: 9rem; }
@media (min-width: 64rem) {
  .panel-tabs { display: none; }
  .workspace-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 1rem; padding: 1.25rem 2rem; }
  .mobile-hidden { display: block; min-width: 0; }
}
@media (max-width: 38rem) {
  .file-picker { align-items: flex-start; flex-direction: column; }
  .file-button { width: 100%; justify-content: center; }
  .action-bar { align-items: stretch; flex-direction: column; }
  .action-bar .button { width: 100%; }
  .import-shell { padding-bottom: 9rem; }
}
</style>
