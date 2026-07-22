import { describe, expect, it } from 'vitest'

import { parseSearchQuery } from './searchQuery'

describe('parseSearchQuery', () => {
  it('preserves quoted values while separating AND terms', () => {
    expect(parseSearchQuery('rating:safe title:"hello world" filesize>=10')).toEqual([
      'rating:safe', 'title:hello world', 'filesize>=10',
    ])
  })

  it('rejects an unfinished quoted value', () => {
    expect(() => parseSearchQuery('title:"unfinished')).toThrow('Close the quoted search value')
  })
})
