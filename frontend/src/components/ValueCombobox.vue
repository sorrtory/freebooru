<script setup lang="ts">
import { computed, shallowRef } from 'vue'

const props = defineProps<{ modelValue: string | string[]; options: string[]; multiple?: boolean }>()
const emit = defineEmits<{ 'update:modelValue': [value: string | string[]] }>()
const query = shallowRef('')
const open = shallowRef(false)
const selected = computed(() => Array.isArray(props.modelValue) ? props.modelValue : props.modelValue ? [props.modelValue] : [])
const filtered = computed(() => props.options.filter((option) => !selected.value.includes(option) && option.toLowerCase().includes(query.value.toLowerCase())))

function choose(value: string) {
  emit('update:modelValue', props.multiple ? [...selected.value, value] : value)
  query.value = ''
  open.value = Boolean(props.multiple)
}

function remove(value: string) {
  emit('update:modelValue', selected.value.filter((item) => item !== value))
}
</script>

<template>
  <div class="combo" :class="{ 'combo--multiple': multiple }" @focusout="open = false">
    <div v-if="multiple && selected.length" class="combo-chips">
      <button v-for="value in selected" :key="value" type="button" @click="remove(value)">{{ value }} ×</button>
    </div>
    <button v-if="!multiple && selected.length" class="combo-current" type="button" @click="open = !open">{{ selected[0] }} <span>⌄</span></button>
    <input v-else v-model="query" :placeholder="multiple ? 'Find a value' : 'Choose a value'" role="combobox" :aria-expanded="open" @focus="open = true" @input="open = true" @keydown.escape="open = false">
    <div v-if="open" class="combo-options" role="listbox">
      <button v-for="option in filtered" :key="option" type="button" role="option" @mousedown.prevent="choose(option)">{{ option }}</button>
      <p v-if="!filtered.length">No matching values</p>
    </div>
  </div>
</template>

<style scoped>
.combo { position: relative; }
.combo input, .combo-current { width: 100%; min-height: 2.75rem; padding: .65rem .75rem; border: 1px solid var(--line-strong); border-radius: .3rem; color: var(--text); background: var(--surface); text-align: left; }
.combo-current { display: flex; justify-content: space-between; align-items: center; }
.combo-options { position: absolute; z-index: 20; top: calc(100% + .3rem); width: 100%; max-height: 15rem; overflow: auto; padding: .25rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface); box-shadow: var(--shadow); }
.combo--multiple .combo-options { position: static; max-height: 10rem; margin-top: .35rem; box-shadow: none; }
.combo-options button { width: 100%; min-height: 2.75rem; padding: .55rem .7rem; border: 0; border-radius: .25rem; color: var(--text); background: transparent; text-align: left; }
.combo-options button:hover { background: var(--primary-soft); }
.combo-options p { margin: .7rem; color: var(--text-muted); }
.combo-chips { display: flex; flex-wrap: wrap; gap: .35rem; margin-bottom: .4rem; }
.combo-chips button { min-height: 2.25rem; padding: .35rem .6rem; border: 1px solid var(--border); border-radius: 999px; color: var(--text); background: var(--primary-soft); }
</style>
