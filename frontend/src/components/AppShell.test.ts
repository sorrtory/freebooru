import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import AppShell from './AppShell.vue'

describe('AppShell', () => {
  it('does not highlight Overview on a nested collection route', async () => {
    const router = createRouter({ history: createMemoryHistory(), routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/settings', component: { template: '<div />' } },
      { path: '/collections/:collection', component: AppShell, props: true, children: [
        { path: '', component: { template: '<div />' } },
        { path: 'search', component: { template: '<div />' } },
        { path: 'files', component: { template: '<div />' } },
        { path: 'import', component: { template: '<div />' } },
      ] },
    ] })
    await router.push('/collections/main/search')
    await router.isReady()

    const wrapper = mount(AppShell, { props: { collection: 'main' }, global: { plugins: [router], stubs: { CollectionSwitcher: true } } })

    expect(wrapper.get('.overview-link').classes()).not.toContain('router-link-exact-active')
    expect(wrapper.get('a[href="/collections/main/search"]').classes()).toContain('router-link-active')
  })
})
