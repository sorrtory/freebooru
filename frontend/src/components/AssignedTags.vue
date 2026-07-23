<script setup lang="ts">
import { computed } from 'vue'

import type { Assignment, ImportField, Relationship, TagValue } from '../api'
import AssignmentCard from './AssignmentCard.vue'

const props = defineProps<{ missing: ImportField[]; assigned: Assignment[]; fields: ImportField[]; conflicts: Set<string>; conflictEdges: Relationship[]; demands: Relationship[] }>()
const emit = defineEmits<{ apply: [name: string, value: TagValue]; remove: [name: string] }>()
const demandedNames = computed(() => new Set(props.demands.map((edge) => edge.target_tag)))
const standaloneMissing = computed(() => props.missing.filter((field) => !demandedNames.value.has(field.name)))

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
      <AssignmentCard v-for="(field, index) in standaloneMissing" :key="field.name" :field="field" problem :open-request="index === 0" @apply="emit('apply', field.name, $event)" />
    </div>
    <div class="panel-section">
      <h3>Applied <span>{{ assigned.length }}</span></h3>
      <p v-if="!assigned.length" class="empty-state">No tags assigned yet.</p>
      <AssignmentCard v-for="assignment in assigned" :key="assignment.name" :field="fieldFor(assignment, fields)" :value="assignment.value" removable :problem="conflicts.has(assignment.name)" :conflict="conflicts.has(assignment.name)" @apply="emit('apply', assignment.name, $event)" @remove="emit('remove', assignment.name)">
        <div v-for="edge in demands.filter((item) => item.source_tag === assignment.name)" :key="`${edge.source_tag}-${edge.target_tag}-${edge.reason}`" class="demand" data-problem>
          <p><strong>{{ edge.target_tag }}</strong> is required by {{ assignment.name }}</p><small>{{ edge.reason || 'Required by this assignment.' }}</small>
          <AssignmentCard v-if="demandField(edge, fields)" :field="demandField(edge, fields)!" :suggested="demandedValue(edge)" problem @apply="emit('apply', edge.target_tag, $event)" />
        </div>
        <p v-for="edge in conflictEdges.filter((item) => item.source_tag === assignment.name || item.target_tag === assignment.name)" :key="`conflict-${edge.source_tag}-${edge.target_tag}`" class="conflict-note">{{ edge.source_tag }} conflicts with {{ edge.target_tag }}<span v-if="edge.reason"> — {{ edge.reason }}</span></p>
      </AssignmentCard>
    </div>
  </section>
</template>

<style scoped>
.demand { margin: .45rem 0 .25rem 1rem; padding: .75rem; border-left: 3px solid var(--warning); background: color-mix(in srgb, var(--warning) 8%, var(--surface)); }.demand p { margin: 0 0 .25rem; }.demand small { color: var(--muted); }.demand :deep(.tag-editor) { margin-top: .65rem; }
.conflict-note { margin: .4rem 0; padding: .6rem .75rem; border-left: 3px solid var(--danger); color: var(--danger); background: color-mix(in srgb, var(--danger) 8%, var(--surface)); font-size: .82rem; }
</style>
