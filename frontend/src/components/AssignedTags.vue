<script setup lang="ts">
import { shallowRef, watch } from 'vue'

import type { Assignment, ImportField, Relationship, TagValue } from '../api'
import TagField from './TagField.vue'

const props = defineProps<{ missing: ImportField[]; assigned: Assignment[]; fields: ImportField[]; conflicts: Set<string>; conflictEdges: Relationship[]; demands: Relationship[] }>()
const emit = defineEmits<{ apply: [name: string, value: TagValue]; remove: [name: string] }>()
const editing = shallowRef('')
const missingEditing = shallowRef('')

watch(() => props.missing, (missing) => {
  if (!missing.some((field) => field.name === missingEditing.value)) missingEditing.value = missing[0]?.name ?? ''
}, { immediate: true })

function fieldFor(assignment: Assignment, fields: ImportField[]): ImportField {
  return fields.find((field) => field.name === assignment.name) ?? { ...assignment, values: [], value_comments: {} }
}

function demandField(edge: Relationship, fields: ImportField[]): ImportField | undefined {
  return fields.find((field) => field.name === edge.target_tag)
}

function demandedValue(edge: Relationship): TagValue | undefined {
  if (edge.target.is !== undefined) return edge.target.is
  if (edge.target.has.length) return edge.target.has
  return undefined
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
      <details v-for="field in missing" :key="field.name" class="assignment" :open="missingEditing === field.name" data-problem><summary @click.prevent="missingEditing = missingEditing === field.name ? '' : field.name"><span><strong>{{ field.name }}</strong><small>{{ field.type }} / required</small></span><b>Needs value</b></summary><TagField v-if="missingEditing === field.name" compact :field="field" @apply="emit('apply', field.name, $event)" @cancel="missingEditing = ''" /></details>
    </div>
    <div class="panel-section">
      <h3>Applied <span>{{ assigned.length }}</span></h3>
      <p v-if="!assigned.length" class="empty-state">No tags assigned yet.</p>
      <article v-for="assignment in assigned" :key="assignment.name" class="assignment" :class="{ 'assignment--conflict': conflicts.has(assignment.name) }" :data-problem="conflicts.has(assignment.name) ? '' : undefined">
        <TagField v-if="editing === assignment.name" compact :field="fieldFor(assignment, fields)" :value="assignment.value" @apply="emit('apply', assignment.name, $event); editing = ''" @remove="emit('remove', assignment.name); editing = ''" @cancel="editing = ''" />
        <button v-else class="assignment-summary" type="button" @click="editing = assignment.name">
          <span><strong>{{ assignment.name }}</strong><small>{{ assignment.type }}{{ assignment.required ? ' / required' : '' }}</small></span>
          <output>{{ Array.isArray(assignment.value) ? assignment.value.join(', ') : assignment.value }}</output>
          <b v-if="conflicts.has(assignment.name)">Conflict</b>
        </button>
        <div v-for="edge in demands.filter((item) => item.source_tag === assignment.name)" :key="`${edge.source_tag}-${edge.target_tag}-${edge.reason}`" class="demand" data-problem>
          <p><strong>{{ edge.target_tag }}</strong> is required by {{ assignment.name }}</p><small>{{ edge.reason || 'Required by this assignment.' }}</small>
          <TagField v-if="demandField(edge, fields)" compact :field="demandField(edge, fields)!" :suggested="demandedValue(edge)" @apply="emit('apply', edge.target_tag, $event)" />
        </div>
        <p v-for="edge in conflictEdges.filter((item) => item.source_tag === assignment.name || item.target_tag === assignment.name)" :key="`conflict-${edge.source_tag}-${edge.target_tag}`" class="conflict-note">{{ edge.source_tag }} conflicts with {{ edge.target_tag }}<span v-if="edge.reason"> — {{ edge.reason }}</span></p>
      </article>
    </div>
  </section>
</template>

<style scoped>
.assignment { margin-bottom: .5rem; }
.assignment > summary { display: flex; min-height: 3.5rem; align-items: center; justify-content: space-between; gap: .75rem; padding: .7rem .8rem; border: 1px solid var(--line); border-radius: .3rem; cursor: pointer; list-style: none; }.assignment > summary span { display: grid; }.assignment > summary small { color: var(--muted); }.assignment > summary b { color: var(--warning); font-size: .72rem; }.assignment[open] > summary { border-color: var(--primary); border-radius: .3rem .3rem 0 0; }
.assignment-summary { display: grid; width: 100%; min-height: 3.5rem; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: .75rem; padding: .7rem .8rem; border: 1px solid var(--line); border-radius: .3rem; color: var(--text); text-align: left; background: rgb(255 255 255 / 2%); cursor: pointer; }
.assignment-summary:hover { border-color: var(--line-strong); background: rgb(255 255 255 / 4%); }
.assignment-summary span { display: grid; min-width: 0; gap: .2rem; }
.assignment-summary strong { overflow-wrap: anywhere; }
.assignment-summary small { color: var(--muted); font: .65rem var(--mono); text-transform: uppercase; }
.assignment-summary output { max-width: 12rem; overflow: hidden; color: var(--accent); font: .75rem var(--mono); text-overflow: ellipsis; white-space: nowrap; }
.assignment-summary b { grid-column: 1 / -1; color: var(--error); font: .67rem var(--mono); letter-spacing: .08em; text-transform: uppercase; }
.assignment--conflict .assignment-summary { border-color: rgb(255 125 105 / 55%); }
.demand { margin: .45rem 0 .25rem 1rem; padding: .75rem; border-left: 3px solid var(--warning); background: color-mix(in srgb, var(--warning) 8%, var(--surface)); }.demand p { margin: 0 0 .25rem; }.demand small { color: var(--muted); }.demand :deep(.tag-editor) { margin-top: .65rem; }
.conflict-note { margin: .4rem 0; padding: .6rem .75rem; border-left: 3px solid var(--danger); color: var(--danger); background: color-mix(in srgb, var(--danger) 8%, var(--surface)); font-size: .82rem; }
</style>
