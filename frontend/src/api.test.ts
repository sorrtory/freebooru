import { describe, expect, it, vi } from 'vitest'

import { createCollection, evaluateImportDraft, getCollection, getCollections, getCollectionStorages, getCollectionTags, getFiles, getHello, getImportSchema, getStatus, importCollectionTag, importFile } from './api'

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
    const body = { collection: 'Main Archive', fields: [{ name: 'rating', type: 'value', comment: 'Safety rating', values: ['safe'], value_comments: { safe: 'Safe content' }, required: true }] }
    const request = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(body), { status: 200 }))

    await expect(getImportSchema('Main Archive', request)).resolves.toEqual(body)
    expect(request).toHaveBeenCalledWith('/api/v1/collections/Main%20Archive/imports/schema', { headers: { Accept: 'application/json' } })
  })

  it('evaluates typed assignments', async () => {
    const body = { collection: 'main', assignments: [{ name: 'pages', type: 'int', comment: '', value: 3, required: false }], missing_required: [], missing_demands: [], active_conflicts: [], suggestions: [], complete: true }
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

describe('collection API', () => {
  it('lists and describes collections', async () => {
    const list = { default_collection: 'main', collections: [{ name: 'main', is_default: true }] }
    const info = { name: 'main', comment: '', is_default: true, tag_count: 5, required_count: 1, storages: ['default'], file_count: 2, total_size_bytes: 40 }
    const request = vi.fn<typeof fetch>().mockResolvedValueOnce(new Response(JSON.stringify(list), { status: 200 })).mockResolvedValueOnce(new Response(JSON.stringify(info), { status: 200 }))

    await expect(getCollections(request)).resolves.toEqual(list)
    await expect(getCollection('main', request)).resolves.toEqual(info)
    expect(request).toHaveBeenNthCalledWith(2, '/api/v1/collections/main', expect.anything())
  })

  it('creates a named collection', async () => {
    const info = { name: 'art', comment: '', is_default: false, tag_count: 5, required_count: 1, storages: ['default'], file_count: 0, total_size_bytes: 0 }
    const request = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(info), { status: 201 }))

    await expect(createCollection('art', request)).resolves.toEqual(info)
    expect(request).toHaveBeenCalledWith('/api/v1/collections', expect.objectContaining({ method: 'POST', body: JSON.stringify({ name: 'art' }) }))
  })

  it('encodes repeated typed search terms and pagination', async () => {
    const page = { files: [], limit: 24, offset: 24, has_more: false, next_offset: null }
    const request = vi.fn<typeof fetch>().mockResolvedValue(new Response(JSON.stringify(page), { status: 200 }))

    await expect(getFiles('main', ['rating:safe', 'title:hello world'], 24, undefined, request)).resolves.toEqual(page)
    expect(request.mock.calls[0][0]).toBe('/api/v1/collections/main/files?limit=24&offset=24&term=rating%3Asafe&term=title%3Ahello+world')
  })

  it('loads and imports explicit collection resources', async () => {
    const tags = { tags: [{ name: 'rating', type: 'value', comment: '', values: [{ value: 'safe', comment: '' }], required: true, imported: true, system: false, assignment_count: 2 }] }
    const storages = { storages: [{ name: 'default', type: 'local', comment: '', imported: true, file_count: 2, total_size_bytes: 42 }] }
    const request = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(new Response(JSON.stringify(tags), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify(storages), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))

    await expect(getCollectionTags('main', 'rat', true, request)).resolves.toEqual(tags.tags)
    await expect(getCollectionStorages('main', request)).resolves.toEqual(storages.storages)
    await expect(importCollectionTag('main', 'artist', request)).resolves.toBeUndefined()
    expect(request).toHaveBeenNthCalledWith(1, '/api/v1/collections/main/tags?q=rat&required=true', expect.anything())
    expect(request).toHaveBeenNthCalledWith(3, '/api/v1/collections/main/tags/artist/import', expect.objectContaining({ method: 'POST' }))
  })
})
