<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import type { ImportField, TagValue } from '../api'
import AssignmentCard from './AssignmentCard.vue'

const props = defineProps<{ fields: ImportField[]; assigned: Set<string>; selected?: string }>()
const emit = defineEmits<{ apply: [name: string, value: TagValue] }>()
const query = shallowRef('')
const selectedName = shallowRef(props.selected ?? '')

watch(() => props.selected, (value) => { if (value) selectedName.value = value })
const available = computed(() => props.fields.filter((field) => !props.assigned.has(field.name) && field.name.toLocaleLowerCase().includes(query.value.trim().toLocaleLowerCase())))
</script>

<template>
  <section class="panel" aria-labelledby="catalog-title">
    <header class="panel-heading">
      <div><p>03 / Tag catalog</p><h2 id="catalog-title">Add tag</h2></div>
      <span class="count">{{ available.length }}</span>
    </header>
    <div class="catalog-search">
      <label for="tag-search">Search imported tags</label>
      <input id="tag-search" v-model="query" type="search" placeholder="rating, artist, storage…">
    </div>
    <div class="catalog-list" aria-label="Available tags">
      <AssignmentCard v-for="field in available" :key="field.name" :field="field" :open-request="selectedName === field.name" @open="selectedName = field.name" @apply="emit('apply', field.name, $event); selectedName = ''" />
      <p v-if="!available.length" class="empty-state">No matching unassigned tags.</p>
    </div>
  </section>
</template>

<style scoped>
.catalog-search { display: grid; gap: .4rem; padding: 1rem; border-bottom: 1px solid var(--line); }
.catalog-search label { color: var(--muted); font: .68rem var(--mono); letter-spacing: .06em; text-transform: uppercase; }
.catalog-search input { min-height: 2.75rem; padding: .65rem .75rem; border: 1px solid var(--line-strong); border-radius: .25rem; color: var(--text); background: var(--surface); }
.catalog-list { display: grid; gap: .4rem; padding: 1rem; }
</style>
