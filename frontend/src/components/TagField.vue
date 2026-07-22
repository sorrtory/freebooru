<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import type { ImportField, TagValue } from '../api'

const props = defineProps<{ field: ImportField; value?: TagValue; compact?: boolean }>()
const emit = defineEmits<{ apply: [value: TagValue]; cancel: []; remove: [] }>()
const localValue = shallowRef<TagValue>(defaultValue())
const validation = shallowRef('')

watch(() => props.value, () => { localValue.value = defaultValue(); validation.value = '' })

const selectedValues = computed(() => Array.isArray(localValue.value) ? localValue.value : [])

function defaultValue(): TagValue {
  if (props.value !== undefined) return Array.isArray(props.value) ? [...props.value] : props.value
  if (props.field.type === 'bool') return true
  if (props.field.type === 'int') return 0
  if (props.field.type === 'multivalue') return []
  return props.field.values[0] ?? ''
}

function updateText(event: Event) {
  localValue.value = (event.target as HTMLInputElement).value
}

function updateNumber(event: Event) {
  localValue.value = (event.target as HTMLInputElement).valueAsNumber
}

function updateBoolean(event: Event) {
  localValue.value = (event.target as HTMLInputElement).checked
}

function updateSelect(event: Event) {
  const target = event.target as HTMLSelectElement
  localValue.value = props.field.type === 'multivalue'
    ? Array.from(target.selectedOptions, (option) => option.value)
    : target.value
}

function submit() {
  if (props.field.type === 'int' && (!Number.isInteger(localValue.value) || !Number.isFinite(localValue.value))) {
    validation.value = 'Enter a whole number.'
    return
  }
  if (typeof localValue.value === 'string' && localValue.value.trim() === '') {
    validation.value = 'Enter a value.'
    return
  }
  if (Array.isArray(localValue.value) && localValue.value.length === 0) {
    validation.value = 'Choose at least one value.'
    return
  }
  validation.value = ''
  emit('apply', localValue.value)
}
</script>

<template>
  <form class="tag-editor" :class="{ 'tag-editor--compact': compact }" @submit.prevent="submit">
    <div class="field-heading">
      <label :for="`tag-${field.name}`">{{ field.name }}</label>
      <span>{{ field.type }}{{ field.required ? ' / required' : '' }}</span>
    </div>
    <label v-if="field.type === 'bool'" class="boolean-control" :for="`tag-${field.name}`">
      <input :id="`tag-${field.name}`" type="checkbox" :checked="Boolean(localValue)" @change="updateBoolean">
      Assign this tag
    </label>
    <select v-else-if="field.type === 'value' || field.type === 'multivalue'" :id="`tag-${field.name}`" :multiple="field.type === 'multivalue'" :value="field.type === 'multivalue' ? selectedValues : localValue" @change="updateSelect">
      <option v-for="option in field.values" :key="option" :value="option">{{ option }}</option>
    </select>
    <input v-else-if="field.type === 'int'" :id="`tag-${field.name}`" type="number" step="1" inputmode="numeric" :value="localValue" @input="updateNumber">
    <input v-else :id="`tag-${field.name}`" :type="field.type === 'datetime' ? 'datetime-local' : field.type === 'date' ? 'date' : 'text'" :value="localValue" @input="updateText">
    <p v-if="validation" class="field-error" role="alert">{{ validation }}</p>
    <div class="editor-actions">
      <button class="button button--primary" type="submit">Apply</button>
      <button v-if="value !== undefined" class="button" type="button" @click="emit('remove')">Remove</button>
      <button v-if="compact" class="button" type="button" @click="emit('cancel')">Cancel</button>
    </div>
  </form>
</template>

<style scoped>
.tag-editor { display: grid; gap: .75rem; padding: 1rem; border: 1px solid var(--line-strong); border-radius: .35rem; background: var(--panel-raised); }
.field-heading { display: flex; align-items: baseline; justify-content: space-between; gap: .75rem; }
.field-heading label { font-weight: 700; overflow-wrap: anywhere; }
.field-heading span { color: var(--muted); font: .68rem var(--mono); letter-spacing: .06em; text-transform: uppercase; }
input:not([type='checkbox']), select { width: 100%; min-height: 2.75rem; padding: .65rem .75rem; border: 1px solid var(--line-strong); border-radius: .25rem; color: var(--text); background: #111411; }
select[multiple] { min-height: 7rem; }
.boolean-control { display: flex; align-items: center; gap: .7rem; min-height: 2.75rem; }
.boolean-control input { width: 1.25rem; height: 1.25rem; accent-color: var(--accent); }
.editor-actions { display: flex; flex-wrap: wrap; gap: .5rem; }
.field-error { margin: 0; color: #ffc1b8; font-size: .86rem; }
</style>
