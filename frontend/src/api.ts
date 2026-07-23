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

export interface CollectionSummary {
  name: string
  is_default: boolean
}

export interface CollectionsResponse {
  default_collection: string
  collections: CollectionSummary[]
}

export interface CollectionInfo extends CollectionSummary {
  comment: string
  tag_count: number
  required_count: number
  storages: string[]
  file_count: number
  total_size_bytes: number
}

export interface FileAssignment {
  name: string
  type: TagType
  value: TagValue
}

export interface FileSource {
  filename: string
  observed_at: string
}

export interface FileRecord {
  sha256: string
  filename: string
  size_bytes: number
  mime_type: string
  imported_at: string
  updated_at: string
  last_interaction_at: string
  assignments: FileAssignment[]
  storages: string[]
  sources: FileSource[]
  content_url: string
}

export interface FilePage {
  files: FileRecord[]
  limit: number
  offset: number
  has_more: boolean
  next_offset: number | null
}

export interface CollectionTag {
  name: string
  type: TagType
  comment: string
  groups: string[]
  values: { value: string; comment: string }[]
  required: boolean
  imported: boolean
  system: boolean
  assignment_count: number
}

export interface CollectionStorage {
  name: string
  type: string
  comment: string
  imported: boolean
  file_count: number
  total_size_bytes: number
}

export interface ApplicationSettings {
  revision: string
  language: 'en' | 'ru'
  default_collection: string
  default_storage_name: string
  http_port: number
  remove_on_upload: boolean
  restart_required: boolean
  collections: string[]
  storages: string[]
}

export type TagType = 'bool' | 'text' | 'int' | 'date' | 'datetime' | 'value' | 'multivalue'
export type TagValue = boolean | number | string | string[]

export interface ImportField {
  name: string
  type: TagType
  comment: string
  values: string[]
  value_comments: Record<string, string>
  required: boolean
}

export interface ImportSchema {
  collection: string
  fields: ImportField[]
}

export interface Assignment extends Omit<ImportField, 'values' | 'value_comments'> {
  value: TagValue
}

export interface Predicate {
  presence: boolean
  has: string[]
  is?: TagValue
  not: string[]
  min?: number
  max?: number
  before?: string
  after?: string
  regex?: string
}

export interface Relationship {
  kind: string
  source_tag: string
  source_value: string
  target_tag: string
  target: Predicate
  reason: string
}

export interface ImportDraft {
  collection: string
  assignments: Assignment[]
  missing_required: ImportField[]
  missing_demands: Relationship[]
  active_conflicts: Relationship[]
  suggestions: Relationship[]
  complete: boolean
}

export interface ImportResult {
  sha256: string
  size_bytes: number
  storages: string[]
  record_created: boolean
  created_copies: string[]
}

export class APIError extends Error {
  constructor(public readonly code: string, message: string) {
    super(message)
    this.name = 'APIError'
  }
}

interface ErrorResponse {
  error?: string | { code?: string; message?: string }
}

export async function getHello(request: typeof fetch = fetch): Promise<HelloResponse> {
  const response = await request('/api/v1/hello', {
    headers: { Accept: 'application/json' },
  })

  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ErrorResponse
    throw new Error(errorMessage(body, response.status))
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
    throw new Error(errorMessage(body, response.status))
  }

  const body: unknown = await response.json()
  if (!isApplicationStatus(body)) {
    throw new Error('FreeBooru returned an invalid status response')
  }
  return body
}

export async function getImportSchema(
  collection: string,
  request: typeof fetch = fetch,
): Promise<ImportSchema> {
  const response = await request(importURL(collection, 'schema'), {
    headers: { Accept: 'application/json' },
  })
  if (!response.ok) throw new Error(errorMessage(await errorBody(response), response.status))
  const body: unknown = await response.json()
  if (!isImportSchema(body)) throw new Error('FreeBooru returned an invalid import schema')
  return body
}

export async function getCollections(request: typeof fetch = fetch): Promise<CollectionsResponse> {
  return requestJSON('/api/v1/collections', {}, isCollectionsResponse, 'collections', request)
}

export async function getCollection(collection: string, request: typeof fetch = fetch): Promise<CollectionInfo> {
  return requestJSON(`/api/v1/collections/${encodeURIComponent(collection)}`, {}, isCollectionInfo, 'collection', request)
}

export async function createCollection(name: string, request: typeof fetch = fetch): Promise<CollectionInfo> {
  return requestJSON('/api/v1/collections', {
    method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ name }),
  }, isCollectionInfo, 'collection', request)
}

export async function getFiles(
  collection: string,
  terms: string[] = [],
  offset = 0,
  signal?: AbortSignal,
  request: typeof fetch = fetch,
): Promise<FilePage> {
  const query = new URLSearchParams({ limit: '24', offset: String(offset) })
  for (const term of terms) query.append('term', term)
  return requestJSON(`/api/v1/collections/${encodeURIComponent(collection)}/files?${query}`, { signal }, isFilePage, 'file results', request)
}

export async function getFile(collection: string, sha256: string, signal?: AbortSignal, request: typeof fetch = fetch): Promise<FileRecord> {
  return requestJSON(`/api/v1/collections/${encodeURIComponent(collection)}/files/${encodeURIComponent(sha256)}`, { signal }, isFileRecord, 'file', request)
}

export async function getCollectionTags(collection: string, query = '', required = false, request: typeof fetch = fetch): Promise<CollectionTag[]> {
  const search = new URLSearchParams()
  if (query) search.set('q', query)
  if (required) search.set('required', 'true')
  const suffix = search.size ? `?${search}` : ''
  const body = await requestJSON(`/api/v1/collections/${encodeURIComponent(collection)}/tags${suffix}`, {}, isCollectionTagsResponse, 'collection tags', request)
  return body.tags
}

export async function importCollectionTag(collection: string, tag: string, request: typeof fetch = fetch): Promise<void> {
  await requestEmpty(`/api/v1/collections/${encodeURIComponent(collection)}/tags/${encodeURIComponent(tag)}/import`, request)
}

export async function getCollectionStorages(collection: string, request: typeof fetch = fetch): Promise<CollectionStorage[]> {
  const body = await requestJSON(`/api/v1/collections/${encodeURIComponent(collection)}/storages`, {}, isCollectionStoragesResponse, 'collection storage', request)
  return body.storages
}

export async function importCollectionStorage(collection: string, storage: string, request: typeof fetch = fetch): Promise<void> {
  await requestEmpty(`/api/v1/collections/${encodeURIComponent(collection)}/storages/${encodeURIComponent(storage)}/import`, request)
}

async function requestEmpty(url: string, request: typeof fetch) {
  const response = await request(url, { method: 'POST', headers: { Accept: 'application/json' } })
  if (!response.ok) throw apiError(await errorBody(response), response.status)
}

export async function getSettings(request: typeof fetch = fetch): Promise<ApplicationSettings> {
  return requestJSON('/api/v1/settings', {}, isApplicationSettings, 'settings', request)
}

export async function updateSettings(settings: ApplicationSettings, request: typeof fetch = fetch): Promise<ApplicationSettings> {
  const { restart_required: _restart, collections: _collections, storages: _storages, ...body } = settings
  return requestJSON('/api/v1/settings', {
    method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
  }, isApplicationSettings, 'settings', request)
}

export async function setFileTag(collection: string, sha256: string, tag: string, value: TagValue, request: typeof fetch = fetch): Promise<FileRecord> {
  return requestJSON(fileTagURL(collection, sha256, tag), {
    method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ value }),
  }, isFileRecord, 'file', request)
}

export async function removeFileTag(collection: string, sha256: string, tag: string, request: typeof fetch = fetch): Promise<FileRecord> {
  return requestJSON(fileTagURL(collection, sha256, tag), { method: 'DELETE' }, isFileRecord, 'file', request)
}

function fileTagURL(collection: string, sha256: string, tag: string) {
  return `/api/v1/collections/${encodeURIComponent(collection)}/files/${encodeURIComponent(sha256)}/tags/${encodeURIComponent(tag)}`
}

export async function evaluateImportDraft(
  collection: string,
  assignments: Record<string, TagValue>,
  signal?: AbortSignal,
  request: typeof fetch = fetch,
): Promise<ImportDraft> {
  const response = await request(importURL(collection, 'evaluate'), {
    method: 'POST',
    headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
    body: JSON.stringify({ assignments }),
    signal,
  })
  if (!response.ok) throw new Error(errorMessage(await errorBody(response), response.status))
  const body: unknown = await response.json()
  if (!isImportDraft(body)) throw new Error('FreeBooru returned an invalid import draft')
  return body
}

export async function importFile(
  collection: string,
  file: File,
  assignments: Record<string, TagValue>,
  signal?: AbortSignal,
  request: typeof fetch = fetch,
): Promise<ImportResult> {
  const form = new FormData()
  form.append('file', file, file.name)
  form.append('assignments', JSON.stringify(assignments))
  const response = await request(`/api/v1/collections/${encodeURIComponent(collection)}/imports`, {
    method: 'POST', headers: { Accept: 'application/json' }, body: form, signal,
  })
  if (!response.ok) throw apiError(await errorBody(response), response.status)
  const body: unknown = await response.json()
  if (!isImportResult(body)) throw new Error('FreeBooru returned an invalid import result')
  return body
}

function importURL(collection: string, action: 'schema' | 'evaluate') {
  return `/api/v1/collections/${encodeURIComponent(collection)}/imports/${action}`
}

async function requestJSON<T>(
  url: string,
  init: RequestInit,
  validate: (value: unknown) => value is T,
  label: string,
  request: typeof fetch,
): Promise<T> {
  const response = await request(url, { ...init, headers: { Accept: 'application/json', ...init.headers } })
  if (!response.ok) throw apiError(await errorBody(response), response.status)
  const body: unknown = await response.json()
  if (!validate(body)) throw new Error(`FreeBooru returned invalid ${label} data`)
  return body
}

async function errorBody(response: Response): Promise<ErrorResponse> {
  return (await response.json().catch(() => ({}))) as ErrorResponse
}

function errorMessage(body: ErrorResponse, status: number): string {
  if (typeof body.error === 'string') return body.error
  if (body.error && typeof body.error.message === 'string') return body.error.message
  return `FreeBooru returned HTTP ${status}`
}

function apiError(body: ErrorResponse, status: number): APIError {
  if (body.error && typeof body.error === 'object') {
    return new APIError(body.error.code ?? 'application.error', body.error.message ?? `FreeBooru returned HTTP ${status}`)
  }
  return new APIError('application.error', errorMessage(body, status))
}

function isImportSchema(value: unknown): value is ImportSchema {
  return isRecord(value) && typeof value.collection === 'string' && Array.isArray(value.fields) && value.fields.every(isImportField)
}

function isImportField(value: unknown): value is ImportField {
  return isRecord(value) && typeof value.name === 'string' && isTagType(value.type) && typeof value.comment === 'string' && Array.isArray(value.values) && value.values.every((item) => typeof item === 'string') && isStringRecord(value.value_comments) && typeof value.required === 'boolean'
}

function isImportDraft(value: unknown): value is ImportDraft {
  return isRecord(value) && typeof value.collection === 'string' && Array.isArray(value.assignments) && value.assignments.every(isAssignment) && Array.isArray(value.missing_required) && value.missing_required.every(isImportField) && Array.isArray(value.missing_demands) && value.missing_demands.every(isRelationship) && Array.isArray(value.active_conflicts) && value.active_conflicts.every(isRelationship) && Array.isArray(value.suggestions) && value.suggestions.every(isRelationship) && typeof value.complete === 'boolean'
}

function isAssignment(value: unknown): value is Assignment {
  return isRecord(value) && typeof value.name === 'string' && isTagType(value.type) && typeof value.comment === 'string' && typeof value.required === 'boolean' && isTagValue(value.value)
}

function isStringRecord(value: unknown): value is Record<string, string> {
  return isRecord(value) && Object.values(value).every((item) => typeof item === 'string')
}

function isRelationship(value: unknown): value is Relationship {
  return isRecord(value) && typeof value.kind === 'string' && typeof value.source_tag === 'string' && typeof value.source_value === 'string' && typeof value.target_tag === 'string' && isPredicate(value.target) && typeof value.reason === 'string'
}

function isPredicate(value: unknown): value is Predicate {
  return isRecord(value) && typeof value.presence === 'boolean' && Array.isArray(value.has) && value.has.every((item) => typeof item === 'string') && Array.isArray(value.not) && value.not.every((item) => typeof item === 'string')
}

function isTagValue(value: unknown): value is TagValue {
  return typeof value === 'boolean' || typeof value === 'number' || typeof value === 'string' || (Array.isArray(value) && value.every((item) => typeof item === 'string'))
}

function isTagType(value: unknown): value is TagType {
  return ['bool', 'text', 'int', 'date', 'datetime', 'value', 'multivalue'].includes(String(value))
}

function isImportResult(value: unknown): value is ImportResult {
  return isRecord(value) && typeof value.sha256 === 'string' && typeof value.size_bytes === 'number' && Array.isArray(value.storages) && value.storages.every((item) => typeof item === 'string') && typeof value.record_created === 'boolean' && Array.isArray(value.created_copies) && value.created_copies.every((item) => typeof item === 'string')
}

function isCollectionsResponse(value: unknown): value is CollectionsResponse {
  return isRecord(value) && typeof value.default_collection === 'string' && Array.isArray(value.collections) && value.collections.every(isCollectionSummary)
}

function isCollectionSummary(value: unknown): value is CollectionSummary {
  return isRecord(value) && typeof value.name === 'string' && typeof value.is_default === 'boolean'
}

function isCollectionInfo(value: unknown): value is CollectionInfo {
  return isRecord(value) && isCollectionSummary(value) && typeof value.comment === 'string' && typeof value.tag_count === 'number' && typeof value.required_count === 'number' && Array.isArray(value.storages) && value.storages.every((item) => typeof item === 'string') && typeof value.file_count === 'number' && typeof value.total_size_bytes === 'number'
}

function isFilePage(value: unknown): value is FilePage {
  return isRecord(value) && Array.isArray(value.files) && value.files.every(isFileRecord) && typeof value.limit === 'number' && typeof value.offset === 'number' && typeof value.has_more === 'boolean' && (value.next_offset === null || typeof value.next_offset === 'number')
}

function isFileRecord(value: unknown): value is FileRecord {
  return isRecord(value) && typeof value.sha256 === 'string' && typeof value.filename === 'string' && typeof value.size_bytes === 'number' && typeof value.mime_type === 'string' && typeof value.imported_at === 'string' && typeof value.updated_at === 'string' && typeof value.last_interaction_at === 'string' && Array.isArray(value.assignments) && value.assignments.every(isFileAssignment) && Array.isArray(value.storages) && value.storages.every((item) => typeof item === 'string') && Array.isArray(value.sources) && value.sources.every(isFileSource) && typeof value.content_url === 'string'
}

function isFileAssignment(value: unknown): value is FileAssignment {
  return isRecord(value) && typeof value.name === 'string' && isTagType(value.type) && isTagValue(value.value)
}

function isFileSource(value: unknown): value is FileSource {
  return isRecord(value) && typeof value.filename === 'string' && typeof value.observed_at === 'string'
}

function isCollectionTagsResponse(value: unknown): value is { tags: CollectionTag[] } {
  return isRecord(value) && Array.isArray(value.tags) && value.tags.every(isCollectionTag)
}

function isCollectionTag(value: unknown): value is CollectionTag {
  return isRecord(value) && typeof value.name === 'string' && isTagType(value.type) && typeof value.comment === 'string' && Array.isArray(value.values) && value.values.every((item) => isRecord(item) && typeof item.value === 'string' && typeof item.comment === 'string') && typeof value.required === 'boolean' && typeof value.imported === 'boolean' && typeof value.system === 'boolean' && typeof value.assignment_count === 'number'
}

function isCollectionStoragesResponse(value: unknown): value is { storages: CollectionStorage[] } {
  return isRecord(value) && Array.isArray(value.storages) && value.storages.every((item) => isRecord(item) && typeof item.name === 'string' && typeof item.type === 'string' && typeof item.comment === 'string' && typeof item.imported === 'boolean' && typeof item.file_count === 'number' && typeof item.total_size_bytes === 'number')
}

function isApplicationSettings(value: unknown): value is ApplicationSettings {
  return isRecord(value) && typeof value.revision === 'string' && (value.language === 'en' || value.language === 'ru') && typeof value.default_collection === 'string' && typeof value.default_storage_name === 'string' && typeof value.http_port === 'number' && typeof value.remove_on_upload === 'boolean' && typeof value.restart_required === 'boolean' && Array.isArray(value.collections) && value.collections.every((item) => typeof item === 'string') && Array.isArray(value.storages) && value.storages.every((item) => typeof item === 'string')
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
