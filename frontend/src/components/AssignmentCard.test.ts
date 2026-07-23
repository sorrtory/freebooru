import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AssignmentCard from './AssignmentCard.vue'

const field = { name: 'artist', type: 'text' as const, comment: '', values: [], value_comments: {}, required: false }

describe('AssignmentCard', () => {
  it('preserves a draft on passive close and resets it on Cancel', async () => {
    const wrapper = mount(AssignmentCard, { props: { field } })
    await wrapper.get('.assignment-summary').trigger('click')
    await wrapper.get('input[type="text"]').setValue('Konata')
    document.body.dispatchEvent(new Event('pointerdown', { bubbles: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.get('.assignment-editor').isVisible()).toBe(false)

    await wrapper.get('.assignment-summary').trigger('click')
    expect((wrapper.get('input[type="text"]').element as HTMLInputElement).value).toBe('Konata')
    await wrapper.get('.tag-editor').trigger('click')
    expect(wrapper.get('.assignment-editor').isVisible()).toBe(false)
    await wrapper.get('.assignment-summary').trigger('click')
    expect((wrapper.get('input[type="text"]').element as HTMLInputElement).value).toBe('Konata')
    await wrapper.findAll('button').find((button) => button.text() === 'Cancel')!.trigger('click')
    await wrapper.get('.assignment-summary').trigger('click')
    expect((wrapper.get('input[type="text"]').element as HTMLInputElement).value).toBe('')
  })

  it('keeps removal available while collapsed', () => {
    const wrapper = mount(AssignmentCard, { props: { field, value: 'Konata', removable: true } })
    expect(wrapper.get('button[aria-label="Remove artist"]').isVisible()).toBe(true)
  })
})
