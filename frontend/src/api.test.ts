import { describe, expect, it, vi } from 'vitest'

import { getHello, getStatus } from './api'

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

describe('getStatus', () => {
  it('accepts a complete application status', async () => {
    const body = {
      ready: true,
      mode: 'desktop',
      default_collection: 'main',
      diagnostics: [
        {
          severity: 'warning',
          code: 'tag.unused',
          message: 'Tag is not imported',
          file: '/config/tags/example.yaml',
          document: 2,
          field: 'name',
        },
      ],
    }
    const request = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify(body), { status: 200 }),
    )

    await expect(getStatus(request)).resolves.toEqual(body)
    expect(request).toHaveBeenCalledWith('/api/v1/status', {
      headers: { Accept: 'application/json' },
    })
  })

  it('reports a JSON API error', async () => {
    const request = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ error: 'temporarily unavailable' }), { status: 503 }),
    )

    await expect(getStatus(request)).rejects.toThrow('temporarily unavailable')
  })

  it.each([
    { ready: true, mode: 'worker', default_collection: 'main', diagnostics: [] },
    { ready: true, mode: 'server', default_collection: 'main' },
    {
      ready: false,
      mode: 'server',
      default_collection: '',
      diagnostics: [{ severity: 'error', code: 'bad' }],
    },
  ])('rejects an invalid status response', async (body) => {
    const request = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify(body), { status: 200 }),
    )

    await expect(getStatus(request)).rejects.toThrow('invalid status response')
  })
})
