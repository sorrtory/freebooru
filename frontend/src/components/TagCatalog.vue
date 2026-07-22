<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import type { ImportField, TagValue } from '../api'
import TagField from './TagField.vue'

const props = defineProps<{ fields: ImportField[]; assigned: Set<string>; selected?: string }>()
const emit = defineEmits<{ apply: [name: string, value: TagValue] }>()
const query = shallowRef('')
const selectedName = shallowRef(props.selected ?? '')

watch(() => props.selected, (value) => { if (value) selectedName.value = value })
const available = computed(() => props.fields.filter((field) => !props.assigned.has(field.name) && field.name.toLocaleLowerCase().includes(query.value.trim().toLocaleLowerCase())))
const selectedField = computed(() => props.fields.find((field) => field.name === selectedName.value))
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
    <TagField v-if="selectedField && !assigned.has(selectedField.name)" :field="selectedField" @apply="emit('apply', selectedField.name, $event); selectedName = ''" />
    <div class="catalog-list" aria-label="Available tags">
      <button v-for="field in available" :key="field.name" type="button" :aria-pressed="selectedName === field.name" @click="selectedName = field.name">
        <span><strong>{{ field.name }}</strong><small>{{ field.type }}</small></span>
        <b>{{ field.required ? 'required' : 'optional' }}</b>
      </button>
      <p v-if="!available.length" class="empty-state">No matching unassigned tags.</p>
    </div>
  </section>
</template>

<style scoped>
.catalog-search { display: grid; gap: .4rem; padding: 1rem; border-bottom: 1px solid var(--line); }
.catalog-search label { color: var(--muted); font: .68rem var(--mono); letter-spacing: .06em; text-transform: uppercase; }
.catalog-search input { min-height: 2.75rem; padding: .65rem .75rem; border: 1px solid var(--line-strong); border-radius: .25rem; color: var(--text); background: #111411; }
.catalog-list { display: grid; gap: .4rem; padding: 1rem; }
.catalog-list button { display: flex; width: 100%; min-height: 3.25rem; align-items: center; justify-content: space-between; gap: .75rem; padding: .65rem .75rem; border: 1px solid var(--line); border-radius: .25rem; color: var(--text); text-align: left; background: transparent; cursor: pointer; }
.catalog-list button:hover, .catalog-list button[aria-pressed='true'] { border-color: var(--accent); background: rgb(181 240 99 / 5%); }
.catalog-list span { display: grid; gap: .2rem; }
.catalog-list small, .catalog-list b { color: var(--muted); font: .65rem var(--mono); text-transform: uppercase; }
</style>
