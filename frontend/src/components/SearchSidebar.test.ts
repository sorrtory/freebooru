import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import SearchSidebar from './SearchSidebar.vue'

afterEach(() => vi.unstubAllGlobals())

describe('SearchSidebar', () => {
  it('completes tag values on Tab and inserts hot tags', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ tags: [
      { name: 'rating', type: 'value', comment: '', values: [{ value: 'safe', comment: '' }], required: true, imported: true, system: false, assignment_count: 4 },
      { name: 'artist', type: 'text', comment: '', values: [], required: false, imported: true, system: false, assignment_count: 2 },
    ] }), { status: 200 })))
    const wrapper = mount(SearchSidebar, { props: { collection: 'main', modelValue: '', errorMessage: '' } })
    await flushPromises()
    const input = wrapper.get('input')
    await input.setValue('rating:s')
    await input.trigger('keydown', { key: 'Tab' })
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual(['rating:safe'])

    await wrapper.get('section button').trigger('click')
    expect(wrapper.emitted('search')).toHaveLength(1)
  })
})
