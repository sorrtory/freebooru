import { flushPromises, mount } from '@vue/test-utils'
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
        { path: 'files', component: { template: '<div />' } },
        { path: 'tags', component: { template: '<div />' } },
        { path: 'storage', component: { template: '<div />' } },
        { path: 'import', component: { template: '<div />' } },
      ] },
    ] })
    await router.push('/collections/main/files?q=rating%3Asafe')
    await router.isReady()

    const wrapper = mount(AppShell, { props: { collection: 'main' }, global: { plugins: [router], stubs: { CollectionSwitcher: true } } })

    expect(wrapper.get('.overview-link').classes()).not.toContain('router-link-exact-active')
    expect(wrapper.get('a[href="/collections/main/files"]').classes()).toContain('router-link-active')
    expect((wrapper.get('input[aria-label="Search files"]').element as HTMLInputElement).value).toBe('rating:safe')

    await wrapper.get('input[aria-label="Search files"]').setValue('artist:konata')
    await wrapper.get('.header-search').trigger('submit')
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe('/collections/main/files?q=artist:konata')
  })
})
