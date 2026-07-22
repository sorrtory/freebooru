export interface HelloResponse {
  message: string
  mode: 'server' | 'desktop'
}

interface ErrorResponse {
  error?: string
}

export async function getHello(request: typeof fetch = fetch): Promise<HelloResponse> {
  const response = await request('/api/v1/hello', {
    headers: { Accept: 'application/json' },
  })

  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ErrorResponse
    throw new Error(body.error ?? `FreeBooru returned HTTP ${response.status}`)
  }

  const body = (await response.json()) as Partial<HelloResponse>
  if (
    body.message !== 'Hello FreeBooru' ||
    (body.mode !== 'server' && body.mode !== 'desktop')
  ) {
    throw new Error('FreeBooru returned an invalid hello response')
  }

  return body as HelloResponse
}
