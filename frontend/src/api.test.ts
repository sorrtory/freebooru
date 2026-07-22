import { describe, expect, it, vi } from 'vitest'

import { evaluateImportDraft, getHello, getImportSchema, getStatus, importFile } from './api'

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

  it('accepts an unready application status', async () => {
    const body = {
      ready: false,
      mode: 'server',
      default_collection: '',
      diagnostics: [
        {
          severity: 'error',
          code: 'application.config_load',
          message: 'Configuration file is missing',
          file: '',
          document: 0,
          field: '',
        },
      ],
    }
    const request = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify(body), { status: 200 }),
    )

    await expect(getStatus(request)).resolves.toEqual(body)
  })

  it('rejects malformed JSON', async () => {
    const request = vi.fn<typeof fetch>().mockResolvedValue(new Response('{', { status: 200 }))

    await expect(getStatus(request)).rejects.toThrow()
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

describe('import workspace API', () => {
  it('loads the explicit collection schema', async () => {
    const body = { collection: 'Main Archive', fields: [{ name: 'rating', type: 'value', values: ['safe'], required: true }] }
    const request = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(body), { status: 200 }))

    await expect(getImportSchema('Main Archive', request)).resolves.toEqual(body)
    expect(request).toHaveBeenCalledWith('/api/v1/collections/Main%20Archive/imports/schema', { headers: { Accept: 'application/json' } })
  })

  it('evaluates typed assignments', async () => {
    const body = { collection: 'main', assignments: [{ name: 'pages', type: 'int', value: 3, required: false }], missing_required: [], missing_demands: [], active_conflicts: [], suggestions: [], complete: true }
    const request = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(body), { status: 200 }))

    await expect(evaluateImportDraft('main', { pages: 3 }, undefined, request)).resolves.toEqual(body)
    expect(request).toHaveBeenCalledWith('/api/v1/collections/main/imports/evaluate', expect.objectContaining({ method: 'POST', body: JSON.stringify({ assignments: { pages: 3 } }) }))
  })

  it('reports structured API errors', async () => {
    const request = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify({ error: { code: 'tag.value_invalid', message: 'pages must be an integer' } }), { status: 400 }))

    await expect(evaluateImportDraft('main', { pages: 3.5 }, undefined, request)).rejects.toThrow('pages must be an integer')
  })

  it('sends file content and typed assignments as multipart data', async () => {
    const result = { sha256: 'abc', size_bytes: 4, storages: ['default'], record_created: true, created_copies: ['default'] }
    const request = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(result), { status: 201 }))
    const file = new File(['data'], 'image.png', { type: 'image/png' })

    await expect(importFile('main', file, { score: 7 }, undefined, request)).resolves.toEqual(result)
    const [, init] = request.mock.calls[0]
    const form = init?.body as FormData
    expect(form.get('file')).toBeInstanceOf(File)
    expect((form.get('file') as File).name).toBe('image.png')
    expect(form.get('assignments')).toBe(JSON.stringify({ score: 7 }))
    expect((init?.headers as Record<string, string>)['Content-Type']).toBeUndefined()
  })
})
