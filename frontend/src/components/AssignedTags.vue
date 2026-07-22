<script setup lang="ts">
import { shallowRef } from 'vue'

import type { Assignment, ImportField, TagValue } from '../api'
import TagField from './TagField.vue'

defineProps<{ missing: ImportField[]; assigned: Assignment[]; fields: ImportField[]; conflicts: Set<string> }>()
const emit = defineEmits<{ apply: [name: string, value: TagValue]; remove: [name: string] }>()
const editing = shallowRef('')

function fieldFor(assignment: Assignment, fields: ImportField[]): ImportField {
  return fields.find((field) => field.name === assignment.name) ?? { ...assignment, values: [] }
}
</script>

<template>
  <section class="panel" aria-labelledby="assigned-title">
    <header class="panel-heading">
      <div><p>01 / Current state</p><h2 id="assigned-title">Assigned</h2></div>
      <span class="count">{{ missing.length + assigned.length }}</span>
    </header>
    <div v-if="missing.length" class="panel-section">
      <h3>Needs value <span>{{ missing.length }}</span></h3>
      <TagField v-for="field in missing" :key="field.name" :field="field" @apply="emit('apply', field.name, $event)" />
    </div>
    <div class="panel-section">
      <h3>Applied <span>{{ assigned.length }}</span></h3>
      <p v-if="!assigned.length" class="empty-state">No tags assigned yet.</p>
      <article v-for="assignment in assigned" :key="assignment.name" class="assignment" :class="{ 'assignment--conflict': conflicts.has(assignment.name) }">
        <TagField v-if="editing === assignment.name" compact :field="fieldFor(assignment, fields)" :value="assignment.value" @apply="emit('apply', assignment.name, $event); editing = ''" @remove="emit('remove', assignment.name); editing = ''" @cancel="editing = ''" />
        <button v-else class="assignment-summary" type="button" @click="editing = assignment.name">
          <span><strong>{{ assignment.name }}</strong><small>{{ assignment.type }}{{ assignment.required ? ' / required' : '' }}</small></span>
          <output>{{ Array.isArray(assignment.value) ? assignment.value.join(', ') : assignment.value }}</output>
          <b v-if="conflicts.has(assignment.name)">Conflict</b>
        </button>
      </article>
    </div>
  </section>
</template>

<style scoped>
.assignment { margin-bottom: .5rem; }
.assignment-summary { display: grid; width: 100%; min-height: 3.5rem; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: .75rem; padding: .7rem .8rem; border: 1px solid var(--line); border-radius: .3rem; color: var(--text); text-align: left; background: rgb(255 255 255 / 2%); cursor: pointer; }
.assignment-summary:hover { border-color: var(--line-strong); background: rgb(255 255 255 / 4%); }
.assignment-summary span { display: grid; min-width: 0; gap: .2rem; }
.assignment-summary strong { overflow-wrap: anywhere; }
.assignment-summary small { color: var(--muted); font: .65rem var(--mono); text-transform: uppercase; }
.assignment-summary output { max-width: 12rem; overflow: hidden; color: var(--accent); font: .75rem var(--mono); text-overflow: ellipsis; white-space: nowrap; }
.assignment-summary b { grid-column: 1 / -1; color: var(--error); font: .67rem var(--mono); letter-spacing: .08em; text-transform: uppercase; }
.assignment--conflict .assignment-summary { border-color: rgb(255 125 105 / 55%); }
</style>
