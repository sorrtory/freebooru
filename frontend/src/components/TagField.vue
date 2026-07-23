<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import type { ImportField, TagValue } from '../api'
import ValueCombobox from './ValueCombobox.vue'

const props = withDefaults(defineProps<{ field: ImportField; value?: TagValue; suggested?: TagValue; compact?: boolean }>(), { value: undefined, suggested: undefined })
const emit = defineEmits<{ apply: [value: TagValue]; cancel: []; dismiss: [] }>()
const localValue = shallowRef<TagValue>(defaultValue())
const validation = shallowRef('')
const choiceValue = computed({
  get: () => localValue.value as string | string[],
  set: (value: string | string[]) => { localValue.value = value },
})

watch(() => props.value, () => { localValue.value = defaultValue(); validation.value = '' })

function defaultValue(): TagValue {
  if (props.value !== undefined) return Array.isArray(props.value) ? [...props.value] : props.value
  if (props.suggested !== undefined) return Array.isArray(props.suggested) ? [...props.suggested] : props.suggested
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
  localValue.value = defaultValue()
}

function cancel() {
  localValue.value = defaultValue()
  validation.value = ''
  emit('cancel')
}
</script>

<template>
  <form class="tag-editor" :class="{ 'tag-editor--compact': compact }" @click.self="emit('dismiss')" @submit.prevent="submit">
    <div class="field-heading">
      <label :for="`tag-${field.name}`">{{ field.name }}</label>
      <span>{{ field.type }}{{ field.required ? ' / required' : '' }}</span>
    </div>
    <p v-if="field.comment" class="field-comment">{{ field.comment }}</p>
    <label v-if="field.type === 'bool'" class="boolean-control" :for="`tag-${field.name}`">
      <input :id="`tag-${field.name}`" type="checkbox" :checked="Boolean(localValue)" @change="updateBoolean">
      Assign this tag
    </label>
    <ValueCombobox v-else-if="field.type === 'value' || field.type === 'multivalue'" v-model="choiceValue" :options="field.values" :multiple="field.type === 'multivalue'" />
    <input v-else-if="field.type === 'int'" :id="`tag-${field.name}`" type="number" step="1" inputmode="numeric" :value="localValue" @input="updateNumber">
    <input v-else :id="`tag-${field.name}`" :type="field.type === 'datetime' ? 'datetime-local' : field.type === 'date' ? 'date' : 'text'" :value="localValue" @input="updateText">
    <p v-if="validation" class="field-error" role="alert">{{ validation }}</p>
    <div class="editor-actions">
      <button class="button button--primary" type="submit">Assign</button>
      <button class="button" type="button" @click="cancel">Cancel</button>
    </div>
  </form>
</template>

<style scoped>
.tag-editor { display: grid; gap: .75rem; padding: 1rem; border: 1px solid var(--line-strong); border-radius: .35rem; background: var(--panel-raised); }
.field-heading { display: flex; align-items: baseline; justify-content: space-between; gap: .75rem; }
.field-comment { margin: 0; color: var(--muted); }
.field-heading label { font-weight: 700; overflow-wrap: anywhere; }
.field-heading span { color: var(--muted); font: .68rem var(--mono); letter-spacing: .06em; text-transform: uppercase; }
input:not([type='checkbox']) { width: 100%; min-height: 2.75rem; padding: .65rem .75rem; border: 1px solid var(--line-strong); border-radius: .25rem; color: var(--text); background: var(--surface); }
.boolean-control { display: flex; align-items: center; gap: .7rem; min-height: 2.75rem; }
.boolean-control input { width: 1.25rem; height: 1.25rem; accent-color: var(--accent); }
.editor-actions { display: flex; flex-wrap: wrap; gap: .5rem; }
.field-error { margin: 0; color: #ffc1b8; font-size: .86rem; }
</style>
