import { describe, expect, it, vi } from 'vitest'

import { getHello } from './api'

describe('getHello', () => {
  it('accepts the server hello response', async () => {
    const request = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ message: 'Hello FreeBooru', mode: 'server' }), {
        headers: { 'Content-Type': 'application/json' },
        status: 200,
      }),
    )

    await expect(getHello(request)).resolves.toEqual({
      message: 'Hello FreeBooru',
      mode: 'server',
    })
    expect(request).toHaveBeenCalledWith('/api/v1/hello', {
      headers: { Accept: 'application/json' },
    })
  })

  it('reports an API error', async () => {
    const request = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ error: 'temporarily unavailable' }), {
        status: 503,
      }),
    )

    await expect(getHello(request)).rejects.toThrow('temporarily unavailable')
  })
})
