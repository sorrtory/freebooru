<script setup lang="ts">
import { onBeforeUnmount, onMounted, shallowRef, useTemplateRef } from 'vue'
import type { editor } from 'monaco-editor/editor/editor.api.js'
import EditorWorker from 'monaco-editor/editor/editor.worker.js?worker'

const props = defineProps<{ url: string; filename: string; mimeType: string; sizeBytes: number }>()
const container = useTemplateRef<HTMLElement>('container')
const loading = shallowRef(true)
const error = shallowRef('')
let instance: editor.IStandaloneCodeEditor | undefined
let request: AbortController | undefined

onMounted(async () => {
  if (props.sizeBytes > 2 * 1024 * 1024) { error.value = 'Text preview is limited to 2 MB.'; loading.value = false; return }
  request = new AbortController()
  try {
    const [response, monaco] = await Promise.all([fetch(props.url, { signal: request.signal }), import('monaco-editor/editor/editor.api.js')])
    if (!response.ok) throw new Error('Text preview is unavailable.')
    const value = await response.text()
    if (!container.value || request.signal.aborted) return
    self.MonacoEnvironment = { getWorker: () => new EditorWorker() }
    instance = monaco.editor.create(container.value, {
      value, language: 'plaintext', readOnly: true, domReadOnly: true,
      automaticLayout: true, minimap: { enabled: false }, scrollBeyondLastLine: false,
      wordWrap: 'on', theme: document.documentElement.dataset.theme === 'dark' ? 'vs-dark' : 'vs',
    })
  } catch (cause) {
    if (!request.signal.aborted) error.value = cause instanceof Error ? cause.message : 'Text preview is unavailable.'
  } finally { loading.value = false }
})
onBeforeUnmount(() => { request?.abort(); instance?.dispose() })
</script>

<template>
  <div class="text-preview"><div class="preview-toolbar"><span>Read only</span><a :href="url">Open file</a></div><p v-if="loading" role="status">Loading text preview…</p><p v-else-if="error" role="alert">{{ error }}</p><div v-show="!loading && !error" ref="container" class="editor" aria-label="Read-only text preview"></div></div>
</template>

<style scoped>
.text-preview { width: 100%; min-height: 24rem; }.preview-toolbar { display: flex; min-height: 2.75rem; align-items: center; justify-content: space-between; padding: .4rem .75rem; border-bottom: 1px solid var(--border); color: var(--text-muted); font-size: .75rem; font-weight: 800; text-transform: uppercase; }.preview-toolbar a { color: var(--primary); }.text-preview > p { margin: 2rem; color: var(--text-muted); }.editor { width: 100%; height: clamp(24rem, 65vh, 52rem); }
</style>
