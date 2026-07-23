<script setup lang="ts">
import { onBeforeUnmount, onMounted, shallowRef, useTemplateRef, watch } from 'vue'

import type { ImportField, TagValue } from '../api'
import TagField from './TagField.vue'

const props = withDefaults(defineProps<{ field: ImportField; value?: TagValue; suggested?: TagValue; removable?: boolean; problem?: boolean; conflict?: boolean; openRequest?: boolean }>(), { value: undefined, suggested: undefined, removable: false, problem: false, conflict: false, openRequest: false })
const emit = defineEmits<{ apply: [value: TagValue]; remove: []; open: [] }>()
const root = useTemplateRef<HTMLElement>('root')
const open = shallowRef(props.openRequest)
const openedOnce = shallowRef(props.openRequest)

watch(() => props.openRequest, (requested) => { if (requested) reveal() })
function reveal() { open.value = true; openedOnce.value = true; emit('open') }
function toggle() { if (open.value) open.value = false; else reveal() }
function apply(value: TagValue) { emit('apply', value); open.value = false }
function closeAndReset() { open.value = false }
function outside(event: PointerEvent) { if (open.value && !root.value?.contains(event.target as Node)) open.value = false }
function display(value: TagValue) { return Array.isArray(value) ? value.join(', ') : String(value) }
onMounted(() => document.addEventListener('pointerdown', outside))
onBeforeUnmount(() => document.removeEventListener('pointerdown', outside))
</script>

<template>
  <article ref="root" class="assignment-card" :class="{ 'assignment-card--open': open, 'assignment-card--conflict': conflict }" :data-problem="problem ? '' : undefined">
    <div class="assignment-summary" @click="toggle">
      <button class="assignment-main" type="button" :aria-expanded="open"><span><strong>{{ field.name }}</strong><small>{{ field.type }}{{ field.required ? ' / required' : '' }}</small></span></button>
      <output v-if="value !== undefined">{{ display(value) }}</output><b v-else>Needs value</b>
      <button v-if="removable" class="trash-button" type="button" :aria-label="`Remove ${field.name}`" title="Remove tag" @click.stop="emit('remove')"><svg aria-hidden="true" viewBox="0 0 24 24"><path d="M4 7h16M9 7V4h6v3m3 0-1 13H7L6 7m4 4v5m4-5v5"/></svg></button>
    </div>
    <div v-if="openedOnce" v-show="open" class="assignment-editor" @click.stop>
      <TagField compact :field="field" :value="value" :suggested="suggested" @apply="apply" @cancel="closeAndReset" @dismiss="open = false" />
    </div>
    <slot />
  </article>
</template>

<style scoped>
.assignment-card { margin-bottom: .5rem; border: 1px solid var(--line); border-radius: .35rem; background: color-mix(in srgb, var(--surface) 97%, var(--primary-soft)); }.assignment-card--open { border-color: var(--primary); }.assignment-card--conflict { border-color: color-mix(in srgb, var(--danger) 60%, var(--border)); }
.assignment-summary { display: grid; min-height: 3.5rem; grid-template-columns: minmax(0, 1fr) auto auto; align-items: center; gap: .55rem; padding: .35rem .45rem .35rem .75rem; cursor: pointer; }.assignment-main { display: block; min-width: 0; min-height: 2.75rem; padding: 0; border: 0; color: var(--text); background: transparent; text-align: left; }.assignment-main span { display: grid; gap: .2rem; }.assignment-main strong { overflow-wrap: anywhere; }.assignment-main small { color: var(--muted); font: .65rem var(--mono); text-transform: uppercase; }.assignment-summary output { max-width: 11rem; overflow: hidden; color: var(--primary); font: .75rem var(--mono); text-overflow: ellipsis; white-space: nowrap; }.assignment-summary > b { color: var(--warning); font-size: .72rem; }
.trash-button { display: grid; width: 2.75rem; min-height: 2.75rem; place-items: center; border: 0; border-radius: var(--radius); color: var(--danger); background: transparent; }.trash-button:hover { background: color-mix(in srgb, var(--danger) 10%, transparent); }.trash-button svg { width: 1.1rem; fill: none; stroke: currentColor; stroke-linecap: round; stroke-linejoin: round; stroke-width: 1.8; }.assignment-editor { border-top: 1px solid var(--line); }.assignment-editor :deep(.tag-editor) { border: 0; border-radius: 0 0 .35rem .35rem; }
</style>
