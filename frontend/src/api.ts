export interface HelloResponse {
  message: string
  mode: 'server' | 'desktop'
}

export type RuntimeMode = 'server' | 'desktop'
export type DiagnosticSeverity = 'error' | 'warning'

export interface Diagnostic {
  severity: DiagnosticSeverity
  code: string
  message: string
  file: string
  document: number
  field: string
}

export interface ApplicationStatus {
  ready: boolean
  mode: RuntimeMode
  default_collection: string
  diagnostics: Diagnostic[]
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

export async function getStatus(request: typeof fetch = fetch): Promise<ApplicationStatus> {
  const response = await request('/api/v1/status', {
    headers: { Accept: 'application/json' },
  })

  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ErrorResponse
    throw new Error(body.error ?? `FreeBooru returned HTTP ${response.status}`)
  }

  const body: unknown = await response.json()
  if (!isApplicationStatus(body)) {
    throw new Error('FreeBooru returned an invalid status response')
  }
  return body
}

function isApplicationStatus(value: unknown): value is ApplicationStatus {
  if (!isRecord(value)) return false
  return (
    typeof value.ready === 'boolean' &&
    (value.mode === 'server' || value.mode === 'desktop') &&
    typeof value.default_collection === 'string' &&
    Array.isArray(value.diagnostics) &&
    value.diagnostics.every(isDiagnostic)
  )
}

function isDiagnostic(value: unknown): value is Diagnostic {
  if (!isRecord(value)) return false
  return (
    (value.severity === 'error' || value.severity === 'warning') &&
    typeof value.code === 'string' &&
    typeof value.message === 'string' &&
    typeof value.file === 'string' &&
    typeof value.document === 'number' &&
    Number.isInteger(value.document) &&
    typeof value.field === 'string'
  )
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
