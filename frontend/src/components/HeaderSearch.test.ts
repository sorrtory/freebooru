import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import HeaderSearch from './HeaderSearch.vue'

afterEach(() => vi.unstubAllGlobals())

describe('HeaderSearch', () => {
  it('shows value completions and Popular tags in one popover', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ tags: [
      { name: 'rating', type: 'value', comment: '', groups: ['general'], values: [{ value: 'safe', comment: '' }], required: true, imported: true, system: false, assignment_count: 4 },
      { name: 'artist', type: 'text', comment: '', groups: ['creator'], values: [{ value: 'studio_trigger', comment: '' }], required: false, imported: true, system: false, assignment_count: 2 },
    ] }), { status: 200 })))
    const wrapper = mount(HeaderSearch, { props: { collection: 'main', modelValue: 'rating:s', 'onUpdate:modelValue': (value: string) => wrapper.setProps({ modelValue: value }) } })
    await flushPromises()
    await wrapper.get('input').trigger('focus')

    expect(wrapper.text()).toContain('rating:safe')
    expect(wrapper.text()).toContain('Popular')
    expect(wrapper.text()).toContain('artist')
    await wrapper.get('input').setValue('artist:trig')
    expect(wrapper.text()).toContain('artist:studio_trigger')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('submit')?.at(-1)).toEqual(['artist:trig'])
  })
})
