import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'

import SettingsPage from '../pages/SettingsPage.vue'

afterEach(() => {
  vi.unstubAllGlobals()
  localStorage.clear()
  document.documentElement.dataset.theme = 'light'
})

describe('Settings appearance', () => {
  it('persists and applies an explicit dark theme', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('offline')))
    const router = createRouter({ history: createMemoryHistory(), routes: [{ path: '/settings', component: SettingsPage }, { path: '/', component: { template: '<div />' } }] })
    await router.push('/settings')
    await router.isReady()
    const wrapper = mount({ template: '<RouterView />' }, { global: { plugins: [router] } })

    await wrapper.get('input[value="dark"]').setValue(true)

    expect(localStorage.getItem('freebooru.theme')).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })
})
