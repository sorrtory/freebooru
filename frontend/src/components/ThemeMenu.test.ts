import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'

import ThemeMenu from './ThemeMenu.vue'

afterEach(() => {
  localStorage.clear()
  document.documentElement.dataset.theme = 'light'
})

describe('ThemeMenu', () => {
  it('persists and applies an explicit dark theme', async () => {
    const wrapper = mount(ThemeMenu)

    await wrapper.get('input[value="dark"]').setValue(true)

    expect(localStorage.getItem('freebooru.theme')).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })
})
