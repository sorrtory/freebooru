import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'

import SettingsPage from '../pages/SettingsPage.vue'

afterEach(() => {
  localStorage.clear()
  document.documentElement.dataset.theme = 'light'
})

describe('Settings appearance', () => {
  it('persists and applies an explicit dark theme', async () => {
    const wrapper = mount(SettingsPage, { global: { stubs: { RouterLink: true } } })

    await wrapper.get('input[value="dark"]').setValue(true)

    expect(localStorage.getItem('freebooru.theme')).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })
})
