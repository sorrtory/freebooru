<script setup lang="ts">
import type { Relationship } from '../api'

defineProps<{ demands: Relationship[]; suggestions: Relationship[] }>()
const emit = defineEmits<{ choose: [relationship: Relationship] }>()

function condition(edge: Relationship): string {
  if (edge.target.is !== undefined) return `is ${String(edge.target.is)}`
  if (edge.target.has.length) return `includes ${edge.target.has.join(', ')}`
  if (edge.target.presence) return 'must be assigned'
  return 'needs a value'
}
</script>

<template>
  <section class="panel" aria-labelledby="guidance-title">
    <header class="panel-heading">
      <div><p>02 / Core guidance</p><h2 id="guidance-title">Suggested</h2></div>
      <span class="count">{{ demands.length + suggestions.length }}</span>
    </header>
    <div class="panel-section">
      <h3>Required next <span>{{ demands.length }}</span></h3>
      <p v-if="!demands.length" class="empty-state">No relationship demands.</p>
      <article v-for="edge in demands" :key="`${edge.source_tag}-${edge.target_tag}-${edge.reason}`" class="relation relation--demand">
        <p><strong>{{ edge.target_tag }}</strong> {{ condition(edge) }}</p>
        <small>{{ edge.source_tag }} activated this demand</small>
        <blockquote>{{ edge.reason || 'Required by the active tag relationship.' }}</blockquote>
        <button class="text-action" type="button" @click="emit('choose', edge)">Set value →</button>
      </article>
    </div>
    <div class="panel-section">
      <h3>Recommended <span>{{ suggestions.length }}</span></h3>
      <p v-if="!suggestions.length" class="empty-state">No recommendations for this draft.</p>
      <article v-for="edge in suggestions" :key="`${edge.source_tag}-${edge.target_tag}-${edge.reason}`" class="relation">
        <p><strong>{{ edge.target_tag }}</strong> {{ condition(edge) }}</p>
        <small>Suggested by {{ edge.source_tag }}</small>
        <blockquote>{{ edge.reason || 'Recommended by the active tag relationship.' }}</blockquote>
        <button class="text-action" type="button" @click="emit('choose', edge)">Review tag →</button>
      </article>
    </div>
  </section>
</template>

<style scoped>
.relation { margin-bottom: .65rem; padding: .85rem; border: 1px dashed rgb(181 240 99 / 34%); border-radius: .3rem; background: rgb(181 240 99 / 3%); }
.relation--demand { border-style: solid; border-color: rgb(243 199 109 / 48%); background: rgb(243 199 109 / 4%); }
.relation p { margin: 0 0 .35rem; }
.relation small { color: var(--muted); font: .68rem var(--mono); }
.relation blockquote { margin: .75rem 0; padding-left: .7rem; border-left: 2px solid var(--line-strong); color: var(--muted-light); font-size: .87rem; line-height: 1.45; }
.text-action { min-height: 2.75rem; padding: 0; border: 0; color: var(--accent); background: transparent; cursor: pointer; font: 700 .72rem var(--mono); text-transform: uppercase; }
</style>
