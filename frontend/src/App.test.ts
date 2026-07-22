import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import App from './App.vue'

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('App', () => {
  it('shows the connected runtime', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(JSON.stringify({ message: 'Hello FreeBooru', mode: 'desktop' }), {
          status: 200,
        }),
      ),
    )

    const wrapper = mount(App)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Hello FreeBooru')
    expect(wrapper.text()).toContain('Wails desktop')
  })

  it('shows an error and retries', async () => {
    const request = vi
      .fn()
      .mockRejectedValueOnce(new Error('network unavailable'))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ message: 'Hello FreeBooru', mode: 'server' }), {
          status: 200,
        }),
      )
    vi.stubGlobal('fetch', request)

    const wrapper = mount(App)
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('network unavailable')

    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.get('h1').text()).toBe('Hello FreeBooru')
    expect(request).toHaveBeenCalledTimes(2)
  })
})
