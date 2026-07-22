import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import ImportPage from './ImportPage.vue'

const schema = {
  collection: 'main',
  fields: [
    { name: 'rating', type: 'value', comment: 'Content safety rating', values: ['safe', 'questionable'], value_comments: { safe: 'Safe content' }, required: true },
    { name: 'featured', type: 'bool', comment: '', values: [], value_comments: {}, required: false },
  ],
}
const emptyDraft = {
  collection: 'main',
  assignments: [],
  missing_required: [schema.fields[0]],
  missing_demands: [],
  active_conflicts: [],
  suggestions: [],
  complete: false,
}

function json(body: unknown) {
  return new Response(JSON.stringify(body), { status: 200 })
}

afterEach(() => vi.unstubAllGlobals())

describe('ImportPage', () => {
  it('loads schema and authoritative draft state', async () => {
    const request = vi.fn().mockResolvedValueOnce(json(schema)).mockResolvedValueOnce(json(emptyDraft))
    vi.stubGlobal('fetch', request)
    const wrapper = mount(ImportPage, { props: { collection: 'main' }, global: { stubs: { RouterLink: true } } })
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Prepare one file')
    expect(wrapper.text()).toContain('Needs value')
    expect(wrapper.text()).toContain('rating')
    expect(wrapper.text()).toContain('1 blocking issue')
    expect(request).toHaveBeenNthCalledWith(1, '/api/v1/collections/main/imports/schema', expect.anything())
  })

  it('applies a typed tag and renders the canonical assignment', async () => {
    const applied = { ...emptyDraft, assignments: [{ name: 'rating', type: 'value', comment: 'Content safety rating', value: 'safe', required: true }], missing_required: [], complete: true }
    const request = vi.fn().mockResolvedValueOnce(json(schema)).mockResolvedValueOnce(json(emptyDraft)).mockResolvedValueOnce(json(applied))
    vi.stubGlobal('fetch', request)
    const wrapper = mount(ImportPage, { props: { collection: 'main' }, global: { stubs: { RouterLink: true } } })
    await flushPromises()
    await wrapper.get('form select').setValue('safe')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('Draft valid')
    expect(wrapper.get('.assignment-summary output').text()).toBe('safe')
    expect(request).toHaveBeenNthCalledWith(3, '/api/v1/collections/main/imports/evaluate', expect.objectContaining({ body: JSON.stringify({ assignments: { rating: 'safe' } }) }))
  })

  it('submits the selected file after Core marks the draft complete', async () => {
    const applied = { ...emptyDraft, assignments: [{ name: 'rating', type: 'value', comment: 'Content safety rating', value: 'safe', required: true }], missing_required: [], complete: true }
    const result = { sha256: 'abcdef1234567890', size_bytes: 4, storages: ['default'], record_created: true, created_copies: ['default'] }
    const request = vi.fn().mockResolvedValueOnce(json(schema)).mockResolvedValueOnce(json(applied)).mockResolvedValueOnce(new Response(JSON.stringify(result), { status: 201 }))
    vi.stubGlobal('fetch', request)
    vi.stubGlobal('URL', { createObjectURL: vi.fn(() => 'blob:preview'), revokeObjectURL: vi.fn() })
    const wrapper = mount(ImportPage, { props: { collection: 'main' }, global: { stubs: { RouterLink: true } } })
    await flushPromises()
    const input = wrapper.get<HTMLInputElement>('#import-file')
    const file = new File(['data'], 'sample.txt', { type: 'text/plain' })
    Object.defineProperty(input.element, 'files', { value: [file] })
    await input.trigger('change')
    await wrapper.get('.action-bar button').trigger('click')
    await flushPromises()

    expect(wrapper.text()).toContain('Imported abcdef123456')
    expect(request).toHaveBeenNthCalledWith(3, '/api/v1/collections/main/imports', expect.objectContaining({ method: 'POST' }))
  })
})
