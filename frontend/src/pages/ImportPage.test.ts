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

  it('focuses the first blocking tag when Import is requested', async () => {
    const request = vi.fn().mockResolvedValueOnce(json(schema)).mockResolvedValueOnce(json(emptyDraft))
    vi.stubGlobal('fetch', request)
    vi.stubGlobal('URL', { createObjectURL: vi.fn(() => 'blob:preview'), revokeObjectURL: vi.fn() })
    vi.stubGlobal('requestAnimationFrame', (callback: FrameRequestCallback) => { callback(0); return 1 })
    Element.prototype.scrollIntoView = vi.fn()
    const wrapper = mount(ImportPage, { props: { collection: 'main' }, attachTo: document.body, global: { stubs: { RouterLink: true } } })
    await flushPromises()
    const input = wrapper.get<HTMLInputElement>('#import-file')
    Object.defineProperty(input.element, 'files', { value: [new File(['data'], 'sample.txt')] })
    await input.trigger('change')
    await wrapper.get('.action-bar button').trigger('click')

    expect(wrapper.get('[data-problem]').classes()).toContain('problem-pulse')
    expect(wrapper.get('[data-problem]').element.contains(document.activeElement)).toBe(true)
    wrapper.unmount()
  })

  it('automatically applies deterministic demanded values', async () => {
    const demandedSchema = { collection: 'main', fields: [
      { name: 'character', type: 'multivalue', comment: '', values: ['konata_izumi'], value_comments: {}, required: false },
      { name: 'universe', type: 'multivalue', comment: '', values: ['lucky_star'], value_comments: {}, required: false },
    ] }
    const demand = { kind: 'demand', source_tag: 'character', source_value: 'konata_izumi', target_tag: 'universe', target: { presence: true, has: ['lucky_star'], not: [] }, reason: '' }
    const first = { collection: 'main', assignments: [{ ...demandedSchema.fields[0], value: ['konata_izumi'] }], missing_required: [], missing_demands: [demand], active_conflicts: [], suggestions: [demand], complete: false }
    const resolved = { ...first, assignments: [...first.assignments, { ...demandedSchema.fields[1], value: ['lucky_star'] }], missing_demands: [], suggestions: [], complete: true }
    const request = vi.fn().mockResolvedValueOnce(json(demandedSchema)).mockResolvedValueOnce(json(first)).mockResolvedValueOnce(json(resolved))
    vi.stubGlobal('fetch', request)
    const wrapper = mount(ImportPage, { props: { collection: 'main' }, global: { stubs: { RouterLink: true } } })
    await flushPromises()

    expect(wrapper.get('section[aria-labelledby="guidance-title"] .count').text()).toBe('0')
    expect(wrapper.text()).toContain('lucky_star')
    expect(request).toHaveBeenNthCalledWith(3, '/api/v1/collections/main/imports/evaluate', expect.objectContaining({ body: JSON.stringify({ assignments: { universe: ['lucky_star'] } }) }))
  })
})
