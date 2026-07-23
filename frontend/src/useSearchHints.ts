import { computed, onMounted, shallowRef, type Ref } from 'vue'

import { getCollectionTags, type CollectionTag } from './api'

export function useSearchHints(collection: string, input: Ref<string>) {
  const tags = shallowRef<CollectionTag[]>([])
  const token = computed(() => input.value.slice(input.value.lastIndexOf(' ') + 1).toLowerCase())
  const suggestions = computed(() => {
    if (!token.value) return []
    const colon = token.value.indexOf(':')
    if (colon >= 0) {
      const name = token.value.slice(0, colon)
      const prefix = token.value.slice(colon + 1)
      const tag = tags.value.find((item) => item.name.toLowerCase() === name)
      return tag?.values
        .filter((item) => item.value.toLowerCase().includes(prefix))
        .map((item) => `${tag.name}:${quoteSearchValue(item.value)}`)
        .slice(0, 8) ?? []
    }
    return tags.value
      .filter((tag) => tag.name.toLowerCase().includes(token.value))
      .map((tag) => tag.type === 'bool' ? tag.name : `${tag.name}:`)
      .slice(0, 8)
  })
  const popular = computed(() => [...tags.value]
    .filter((tag) => tag.assignment_count > 0)
    .sort((left, right) => right.assignment_count - left.assignment_count || left.name.localeCompare(right.name))
    .slice(0, 12))

  onMounted(async () => {
    try { tags.value = (await getCollectionTags(collection)).filter((tag) => tag.imported && !tag.system) }
    catch { tags.value = [] }
  })
  return { tags, suggestions, popular }
}

export function quoteSearchValue(value: string) { return /\s/.test(value) ? `"${value}"` : value }

export function replaceSearchToken(input: string, value: string) {
  const boundary = input.lastIndexOf(' ')
  return `${boundary >= 0 ? input.slice(0, boundary + 1) : ''}${value}`
}
