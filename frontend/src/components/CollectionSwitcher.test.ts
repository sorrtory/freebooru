import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, describe, expect, it, vi } from 'vitest'

import CollectionSwitcher from './CollectionSwitcher.vue'

afterEach(() => vi.unstubAllGlobals())

describe('CollectionSwitcher', () => {
  it('closes its popover on an outside pointer interaction', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ default_collection: 'main', collections: [{ name: 'main', is_default: true }] }), { status: 200 })))
    const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/:pathMatch(.*)*', component: { template: '<div />' } }] })
    const wrapper = mount(CollectionSwitcher, { props: { collection: 'main' }, attachTo: document.body, global: { plugins: [router] } })
    await flushPromises()
    await wrapper.get('.switcher-button').trigger('click')
    expect(wrapper.find('.switcher-popover').exists()).toBe(true)
    document.body.dispatchEvent(new Event('pointerdown', { bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.switcher-popover').exists()).toBe(false)
    wrapper.unmount()
  })
})
