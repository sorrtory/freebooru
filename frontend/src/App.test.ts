import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import StatusPage from './pages/StatusPage.vue'

function response(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200 })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

const mountStatus = () => mount(StatusPage, { global: { stubs: { RouterLink: true } } })

describe('StatusPage', () => {
  it('shows a ready desktop application', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        response({
          ready: true,
          mode: 'desktop',
          default_collection: 'main',
          diagnostics: [],
        }),
      ),
    )

    const wrapper = mountStatus()
    expect(wrapper.get('[role="status"]').text()).toContain('Loading')
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('FreeBooru is ready')
    expect(wrapper.text()).toContain('Wails desktop')
    expect(wrapper.text()).toContain('main')
    expect(wrapper.find('button').exists()).toBe(false)
  })

  it('shows warnings while remaining ready', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        response({
          ready: true,
          mode: 'server',
          default_collection: 'main',
          diagnostics: [
            {
              severity: 'warning',
              code: 'tag.unused',
              message: 'Tag is not imported',
              file: '/config/tags/example.yaml',
              document: 2,
              field: 'name',
            },
          ],
        }),
      ),
    )

    const wrapper = mountStatus()
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('FreeBooru is ready')
    expect(wrapper.text()).toContain('Web server')
    expect(wrapper.get('[aria-label="Configuration diagnostics"]').text()).toContain(
      'tag.unused',
    )
    expect(wrapper.text()).toContain('/config/tags/example.yaml · document 2 · name')
    expect(wrapper.text()).toContain('1 diagnostic')
  })

  it('shows configuration errors without empty source metadata', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        response({
          ready: false,
          mode: 'server',
          default_collection: '',
          diagnostics: [
            {
              severity: 'error',
              code: 'application.config_load',
              message: 'Configuration file is missing',
              file: '',
              document: 0,
              field: '',
            },
          ],
        }),
      ),
    )

    const wrapper = mountStatus()
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Configuration needs attention')
    expect(wrapper.text()).toContain('Configuration file is missing')
    expect(wrapper.text()).not.toContain('document 0')
    expect(wrapper.get('button').text()).toBe('Retry')
  })

  it('recovers after retrying a configuration error', async () => {
    const request = vi
      .fn()
      .mockResolvedValueOnce(
        response({
          ready: false,
          mode: 'server',
          default_collection: '',
          diagnostics: [
            {
              severity: 'error',
              code: 'application.config_load',
              message: 'Configuration file is missing',
              file: '',
              document: 0,
              field: '',
            },
          ],
        }),
      )
      .mockResolvedValueOnce(
        response({
          ready: true,
          mode: 'server',
          default_collection: 'main',
          diagnostics: [],
        }),
      )
    vi.stubGlobal('fetch', request)

    const wrapper = mountStatus()
    await flushPromises()
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('FreeBooru is ready')
    expect(request).toHaveBeenCalledTimes(2)
  })

  it('shows a network error and retries', async () => {
    const request = vi
      .fn()
      .mockRejectedValueOnce(new Error('network unavailable'))
      .mockResolvedValueOnce(
        response({
          ready: true,
          mode: 'desktop',
          default_collection: 'main',
          diagnostics: [],
        }),
      )
    vi.stubGlobal('fetch', request)

    const wrapper = mountStatus()
    await flushPromises()
    expect(wrapper.get('[role="alert"]').text()).toContain('network unavailable')

    await wrapper.get('button').trigger('click')
    await flushPromises()
    expect(wrapper.get('h1').text()).toBe('FreeBooru is ready')
  })
})
