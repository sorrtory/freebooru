<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue'
import { RouterLink } from 'vue-router'
import { useRoute } from 'vue-router'

import { getFile, getImportSchema, removeFileTag, setFileTag, type FileRecord, type ImportField, type TagValue } from '../api'
import TagField from '../components/TagField.vue'
import TextPreview from '../components/TextPreview.vue'

const props = defineProps<{ collection: string; sha256: string }>()
const route = useRoute()
const file = shallowRef<FileRecord>()
const fields = shallowRef<ImportField[]>([])
const editing = shallowRef('')
const error = shallowRef('')
const saving = shallowRef(false)
const isImage = computed(() => file.value?.mime_type.startsWith('image/'))
const isText = computed(() => Boolean(file.value && (file.value.mime_type.startsWith('text/') || /(json|javascript|xml|yaml)/i.test(file.value.mime_type) || /\.(md|txt|json|ya?ml|xml|html?|css|[cm]?[jt]sx?)$/i.test(file.value.filename))))
const available = computed(() => fields.value.filter((field) => field.name !== 'storage' && !file.value?.assignments.some((assignment) => assignment.name === field.name)))
const backTo = computed(() => typeof route.query.from === 'string' && route.query.from.startsWith(`/collections/${props.collection}/`) ? route.query.from : `/collections/${props.collection}/files`)

onMounted(load)

async function load() {
  error.value = ''
  try {
    const [record, schema] = await Promise.all([getFile(props.collection, props.sha256), getImportSchema(props.collection)])
    file.value = record
    fields.value = schema.fields
  } catch (cause) { error.value = cause instanceof Error ? cause.message : 'Could not open this file.' }
}

function fieldFor(name: string): ImportField | undefined {
  const field = fields.value.find((item) => item.name === name)
  const assignment = file.value?.assignments.find((item) => item.name === name)
  return field ?? (assignment ? { ...assignment, comment: '', values: [], value_comments: {}, required: false } : undefined)
}

async function save(name: string, value: TagValue) {
  if (!file.value) return
  saving.value = true; error.value = ''
  try { file.value = await setFileTag(props.collection, props.sha256, name, value); editing.value = '' }
  catch (cause) { error.value = cause instanceof Error ? cause.message : 'Could not save the tag.' }
  finally { saving.value = false }
}

async function remove(name: string) {
  if (!file.value) return
  saving.value = true; error.value = ''
  try { file.value = await removeFileTag(props.collection, props.sha256, name); editing.value = '' }
  catch (cause) { error.value = cause instanceof Error ? cause.message : 'Could not remove the tag.' }
  finally { saving.value = false }
}

function formatBytes(bytes: number) { return new Intl.NumberFormat(undefined, { notation: 'compact', style: 'unit', unit: 'byte' }).format(bytes) }
</script>

<template>
  <main class="page detail-page">
    <RouterLink class="back-link" :to="backTo">← Back</RouterLink>
    <p v-if="error" class="detail-error" role="alert">{{ error }}</p>
    <p v-if="!file && !error" class="empty-copy">Loading…</p>
    <template v-if="file">
      <section class="preview-panel">
        <img v-if="isImage" :src="file.content_url" :alt="file.filename">
        <TextPreview v-else-if="isText" :url="file.content_url" :filename="file.filename" :mime-type="file.mime_type" :size-bytes="file.size_bytes" />
        <div v-else class="file-fallback"><strong>{{ file.mime_type || 'File' }}</strong><a class="button" :href="file.content_url">Open file</a></div>
      </section>
      <section class="detail-main">
        <header><h1>{{ file.filename }}</h1><p>{{ formatBytes(file.size_bytes) }} · {{ file.mime_type || 'Unknown type' }}</p></header>
        <div class="meta"><span>Imported {{ new Date(file.imported_at).toLocaleDateString() }}</span><code :title="file.sha256">{{ file.sha256.slice(0, 12) }}…</code><span>{{ file.storages.join(', ') }}</span></div>
        <section class="tag-list">
          <div class="section-heading"><h2>Tags</h2><span>{{ file.assignments.length }}</span></div>
          <article v-for="assignment in file.assignments" :key="assignment.name" class="tag-row">
            <template v-if="editing !== assignment.name"><div><strong>{{ assignment.name }}</strong><span>{{ Array.isArray(assignment.value) ? assignment.value.join(', ') : assignment.value }}</span></div><button class="button" @click="editing = assignment.name">Edit</button></template>
            <TagField v-else-if="fieldFor(assignment.name)" :field="fieldFor(assignment.name)!" :value="assignment.value" compact @apply="save(assignment.name, $event)" @remove="remove(assignment.name)" @cancel="editing = ''" />
          </article>
          <details v-if="available.length" class="add-tag"><summary>Add tag</summary><div class="tag-options"><button v-for="field in available" :key="field.name" class="button" @click="editing = field.name">{{ field.name }}</button></div><TagField v-if="available.some((field) => field.name === editing)" :field="available.find((field) => field.name === editing)!" compact @apply="save(editing, $event)" @cancel="editing = ''" /></details>
        </section>
      </section>
    </template>
  </main>
</template>

<style scoped>
.detail-page { display: grid; align-content: start; gap: 1rem; }
.back-link { color: var(--primary); font-weight: 800; }
.preview-panel { display: grid; min-height: 18rem; max-height: 70vh; place-items: center; overflow: hidden; border: 1px solid var(--border); border-radius: var(--radius-lg); background: var(--surface-raised); }
.preview-panel:has(.text-preview) { max-height: none; }
.preview-panel img { width: 100%; height: 100%; max-height: 70vh; object-fit: contain; }
.file-fallback { display: grid; gap: 1rem; justify-items: center; }
.detail-main { min-width: 0; }
.detail-main header h1 { margin: 0; overflow-wrap: anywhere; font: clamp(1.8rem, 5vw, 3rem)/1.05 var(--display); }
.detail-main header p, .meta { color: var(--text-muted); }
.meta { display: flex; flex-wrap: wrap; gap: .5rem 1rem; margin: 1rem 0 1.5rem; font-size: .85rem; }
.tag-list { border-top: 1px solid var(--border); }
.section-heading, .tag-row > :first-child { display: flex; justify-content: space-between; gap: 1rem; align-items: center; }
.section-heading h2 { margin: 1rem 0; }
.tag-row { padding: .75rem 0; border-top: 1px solid var(--line); }
.tag-row > div span { margin-left: 1rem; color: var(--text-muted); overflow-wrap: anywhere; }
.tag-row :deep(.tag-editor) { width: 100%; }
.add-tag { padding: 1rem 0; border-top: 1px solid var(--border); }
.add-tag summary { min-height: 2.75rem; cursor: pointer; font-weight: 800; }
.tag-options { display: flex; flex-wrap: wrap; gap: .5rem; margin-bottom: 1rem; }
.detail-error { padding: .8rem; border: 1px solid var(--danger); color: var(--danger); background: var(--surface); }
@media (min-width: 64rem) { .detail-page { grid-template-columns: minmax(0, 1.35fr) minmax(20rem, .65fr); } .back-link { grid-column: 1 / -1; } }
</style>
