import { onBeforeUnmount, shallowRef } from 'vue'

import { evaluateImportDraft, getImportSchema } from './api'
import type { ImportDraft, ImportSchema, TagValue } from './api'

export function useImportDraft(collection: string) {
  const schema = shallowRef<ImportSchema>()
  const draft = shallowRef<ImportDraft>()
  const assignments = shallowRef<Record<string, TagValue>>({})
  const loading = shallowRef(true)
  const evaluating = shallowRef(false)
  const errorMessage = shallowRef('')
  let evaluation: AbortController | undefined

  async function load() {
    loading.value = true
    errorMessage.value = ''
    try {
      schema.value = await getImportSchema(collection)
      await evaluate(assignments.value)
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : 'Unable to load import workspace'
    } finally {
      loading.value = false
    }
  }

  async function evaluate(next: Record<string, TagValue>) {
    evaluation?.abort()
    const controller = new AbortController()
    evaluation = controller
    evaluating.value = true
    try {
      let result = await evaluateImportDraft(collection, next, controller.signal)
      if (evaluation !== controller) return
      const automatic = deterministicDemands(result)
      if (Object.keys(automatic).length) {
        result = await evaluateImportDraft(collection, { ...next, ...automatic }, controller.signal)
        if (evaluation !== controller) return
      }
      draft.value = result
      assignments.value = Object.fromEntries(result.assignments.map((item) => [item.name, item.value]))
      errorMessage.value = ''
    } catch (error) {
      if (controller.signal.aborted) return
      errorMessage.value = error instanceof Error ? error.message : 'Unable to evaluate tags'
    } finally {
      if (evaluation === controller) evaluating.value = false
    }
  }

  function deterministicDemands(result: ImportDraft): Record<string, TagValue> {
    const assigned = new Set(result.assignments.map((item) => item.name))
    const values: Record<string, TagValue> = {}
    for (const edge of result.missing_demands) {
      if (assigned.has(edge.target_tag) || values[edge.target_tag] !== undefined) continue
      if (edge.target.is !== undefined) values[edge.target_tag] = edge.target.is
      else if (edge.target.has.length) values[edge.target_tag] = edge.target.has
      else if (edge.target.presence && schema.value?.fields.find((field) => field.name === edge.target_tag)?.type === 'bool') values[edge.target_tag] = true
    }
    return values
  }

  function apply(name: string, value: TagValue) {
    return evaluate({ ...assignments.value, [name]: value })
  }

  function remove(name: string) {
    const next = { ...assignments.value }
    delete next[name]
    return evaluate(next)
  }

  onBeforeUnmount(() => evaluation?.abort())
  return { schema, draft, assignments, loading, evaluating, errorMessage, load, apply, remove }
}
