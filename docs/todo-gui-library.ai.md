# FreeBooru library GUI contract

This document defines the next GUI work after
[todo-gui-core.ai.md](./todo-gui-core.ai.md). It turns the shared Vue
application into a minimal usable library while preserving the same HTTP API
inside the standalone server and Wails.

The contract covers collection selection, paginated search, file details and
content, browser upload, and file tag/storage assignment. Editing YAML tag or
collection definitions is not part of this slice.

## Goal

A user with valid configuration can:

1. select a configured collection;
2. search or browse its files;
3. open a file and view its metadata/content;
4. import one file with typed initial assignments;
5. add, replace, or remove file tag assignments;
6. add or remove storage assignments with explicit destructive confirmation;
7. recover from validation, relationship, stale-edit, and transport errors.

The workflow must behave the same in a browser and in Wails.

```text
Vue route + application state
            |
      relative /api/v1
            |
      webapi DTO mapping
            |
          Core
       /        \
HTTP server    Wails asset middleware
```

## What was missing from the initial feature list

The proposed collection, upload, tag, and search operations also require:

- collection listing and collection metadata for the selector;
- typed field definitions for import and edit forms;
- tag-name completion and allowed-value hints;
- file metadata and a content endpoint so search results can be displayed;
- pagination with stable ordering and a defined maximum page size;
- structured errors rather than UI parsing of Go error strings;
- relationship diagnostics for rejected mutations;
- duplicate-import behavior;
- upload size limits, streaming temporary files, and cleanup;
- storage assignment, which is a special tag with physical side effects;
- last-copy deletion confirmation;
- optimistic concurrency so two tabs cannot silently overwrite each other;
- cancellation and stale-response handling for search and navigation;
- explicit empty, loading, unavailable, and partial-failure UI states.

## Fixed architectural decisions

### Collection selection is client state

“Open collection” means selecting `/collections/:collection` in Vue. It does
not create a process-global HTTP session.

Every collection API request includes the collection name. Core may open a
database for the duration of an operation and close it afterward. The API must
not call `Core.OpenCollection` and retain that session across requests.
This follows the shared [Core contract](./core.ai.md): web and Wails callers
pass an explicit collection on every collection-scoped Core call and never
depend on the empty-name default.

This rule is required because:

- the HTTP server may serve multiple tabs or users;
- a process-global open collection would make different clients contend;
- an open Core session prevents configuration reload/check;
- request-scoped ownership gives cancellation and cleanup a clear boundary.

Wails uses the same stateless API even though it normally has one user.

### Configuration gates library routes

- Vue requests status before entering library routes.
- Library endpoints require a successfully loaded and checked configuration
  snapshot.
- If configuration is unavailable, they return `application.not_ready`.
- Vue returns to the existing configuration status screen and offers Retry.
- Configuration is not reloaded implicitly during every library operation.
- No library operation may run against a status response with `ready: false`.

### “Edit tag” means a file assignment

This slice edits tag values assigned to an indexed file. It does not modify
tag schemas, predefined values, relationships, collection YAML, storage YAML,
or application YAML. Configuration editing requires its own contract with
atomic file writes, validation, and conflict handling.

### One transport contract

- Vue uses only relative `/api/v1/...` requests.
- Browser and Wails uploads both use `multipart/form-data` through the shared
  handler.
- A future native Wails picker may submit a trusted local path through a
  desktop-only adapter, but it is not required for the first library UI.
- Browser clients never submit an arbitrary server filesystem path.

### API models stay independent

HTTP DTOs do not expose `config`, `collection`, `evaluator`, SQLite, or Wails
types. `webapi` maps Core results into explicit JSON fields. Interfaces are
defined by the route group that consumes them rather than one oversized Core
interface.

## UI information architecture

### Routes

| Route | Purpose |
| --- | --- |
| `/` | Status gate and redirect to the configured default collection |
| `/collections` | Configured collection chooser |
| `/collections/:collection` | Searchable file grid |
| `/collections/:collection/files/:sha256` | File content, metadata, tags, and storages |
| `/collections/:collection/import` | One-file import form |

The selected collection and opened file are encoded in the URL. Search terms,
limit, and offset are URL query state so refresh/back/forward remain useful.

### Shared application state

Use Vue Router plus a small application/session composable initially. Do not
add Pinia until independent feature areas genuinely need shared mutable state.

The sources of truth are:

- current route for collection, file, and search parameters;
- status response for application readiness;
- the most recent server response for collection/file data;
- local form state only while a form is being edited.

Derived labels, counts, and button availability use computed state. Search and
navigation cancel superseded requests with `AbortController` and ignore stale
responses.

### Required screens and states

- collection chooser: loading, populated, empty, and unavailable;
- file grid: initial loading, results, empty collection, no matches, next/prev,
  and request error;
- file detail: loading, content preview/fallback, metadata, assignments, and
  not found;
- import: schema loading, validation error, uploading, success, duplicate, and
  transport failure;
- tag editor: type-appropriate input, allowed/unavailable values, relationship
  rejection, saving, stale revision, and success;
- storage editor: copy progress, failure, removal confirmation, and
  file-deleted result.

## Resource model

### Collection summary

```json
{
  "name": "main",
  "is_default": true
}
```

Only valid configured collections are listed. Filesystem/database locations
are not exposed to remote clients.

### File summary and detail

Search summaries contain enough data to render a grid without a second
request per item:

```json
{
  "sha256": "hex-encoded-sha256",
  "size_bytes": 12345,
  "imported_at": "2026-07-22T12:00:00Z",
  "updated_at": "2026-07-22T12:00:00Z",
  "filename": "image.png",
  "tags": [
    { "name": "artist", "type": "text", "value": "someone" }
  ],
  "storages": ["default"],
  "content_url": "/api/v1/collections/main/files/hex-encoded-sha256/content"
}
```

File detail additionally includes all observed sources. Source paths are
sensitive local information: return filenames and timestamps by default, not
absolute source paths. A future trusted-desktop option may expose paths.

Tag values use these JSON shapes:

| Tag type | JSON value |
| --- | --- |
| `bool` | boolean |
| `text`, `date`, `datetime`, `value` | string |
| `int` | integer |
| `multivalue` | string array |
| built-in `storage` | string array through storage endpoints |

Every array is encoded as `[]`, never `null`.

### Revision

File responses include an HTTP `ETag` derived from the persisted file
revision. `updated_at` alone is not a safe opaque revision if its storage
precision permits equal timestamps.

Every tag or storage mutation requires `If-Match`. A stale revision returns
`409 file.revision_conflict` and the UI reloads the file before allowing a new
edit. Implementing this requires conditional persistence in Core/repository;
checking only in `webapi` is racy and is not sufficient.

## HTTP contract

All success responses are JSON unless the content endpoint is used. All route
parameters are URL-escaped and collection/tag matching retains Core’s
case-insensitive rules.

### Collections

```http
GET /api/v1/collections
GET /api/v1/collections/{collection}
```

The list response is:

```json
{
  "default_collection": "main",
  "collections": [{ "name": "main", "is_default": true }]
}
```

The detail request validates that the collection is usable and returns its
summary plus import/edit field definitions. It is the server-side validation
behind selecting/opening a collection in Vue.

### Search and browse

```http
GET /api/v1/collections/{collection}/files?term=artist:someone&term=-draft&limit=50&offset=0
```

- `term` is repeatable and preserves the existing Core query grammar.
- no terms means browse all files;
- default limit is `50`; maximum limit is `100`;
- negative or excessive pagination values return `request.invalid`;
- ordering remains deterministic according to the repository contract;
- the implementation fetches at most `limit + 1` records to calculate
  `has_more` without an expensive total-count query.

```json
{
  "files": [],
  "limit": 50,
  "offset": 0,
  "has_more": false,
  "next_offset": null
}
```

### File metadata

```http
GET /api/v1/collections/{collection}/files/{sha256}
```

The hash must be exactly 64 lowercase hexadecimal characters. Missing files
return `404 file.not_found`.

### File content

```http
GET /api/v1/collections/{collection}/files/{sha256}/content
HEAD /api/v1/collections/{collection}/files/{sha256}/content
```

- Core chooses an assigned available storage copy; the client never supplies
  a filesystem path.
- Set an explicit media type when safely detected, otherwise
  `application/octet-stream`.
- Set `Content-Length`, `ETag`, `X-Content-Type-Options: nosniff`, and a safe
  `Content-Disposition` filename.
- Support standard single-range requests so video and large-file preview does
  not require downloading the entire file.
- Stream content; never read the complete file into memory.
- Inline preview is initially limited to browser-supported image, audio,
  video, PDF, and text types. Other types show metadata and Download.

This requires a new Core content-opening workflow returning metadata and a
seekable/closable reader without exposing the storage path.

### Tag definitions and value hints

```http
GET /api/v1/collections/{collection}/tags?prefix=art&limit=20
GET /api/v1/collections/{collection}/files/{sha256}/tags/{tag}/values
```

Tag search returns imported tag definitions with name, type, predefined
values, and whether the tag is required. Value hints map
`Core.AllowedValues`, including `allowed` and structured reasons. Hints guide
the UI but the mutation endpoint always performs authoritative validation.

### Set or remove a file tag

```http
PUT /api/v1/collections/{collection}/files/{sha256}/tags/{tag}
If-Match: "opaque-file-revision"
Content-Type: application/json

{"value":"someone"}
```

```http
DELETE /api/v1/collections/{collection}/files/{sha256}/tags/{tag}
If-Match: "opaque-file-revision"
```

`PUT` means complete replacement and handles both adding an absent assignment
and editing an existing one. A separate HTTP operation for `Core.AddTag` is not
needed in the first UI. Multivalue editing submits the complete desired array.

Success returns the updated file and new `ETag`, not only `changed: true`.
Relationship rejection returns structured missing-demand and active-conflict
details and does not persist a partial state.

### Add or remove storage

```http
PUT /api/v1/collections/{collection}/files/{sha256}/storages/{storage}
If-Match: "opaque-file-revision"
```

```http
DELETE /api/v1/collections/{collection}/files/{sha256}/storages/{storage}
If-Match: "opaque-file-revision"
```

Storage is separate from ordinary tag endpoints because it copies/deletes
physical content. Removing the final assigned storage can delete the indexed
file. The first request must return
`409 storage.last_copy_confirmation_required` without mutation. The UI shows
an explicit warning and retries with:

```http
X-FreeBooru-Confirm-Delete: true
```

The success response states whether a copy and/or the file record was deleted.

### Import form schema

```http
GET /api/v1/collections/{collection}/imports/schema
```

The response maps `Core.ImportFields` and includes tag name, type, predefined
values, and required state. Vue renders controls by type rather than hardcoded
tag names.

### Browser and Wails upload

```http
POST /api/v1/collections/{collection}/imports
Content-Type: multipart/form-data; boundary=...
```

Multipart parts:

- `file`: exactly one file;
- `assignments`: one UTF-8 JSON object using the typed tag value shapes above.

Rules:

- stream the file into a private temporary file;
- enforce a documented request limit before parsing and return `413` when it
  is exceeded;
- choose the initial limit before implementation and keep it in one named
  server constant (recommended first default: 10 GiB);
- reject multiple file parts and unexpected large form fields;
- preserve the client filename as display metadata after sanitizing it to a
  basename; never use it as a destination path;
- always clean the temporary file after Core succeeds, fails, or the request
  is cancelled;
- extend the Core import input to accept an explicit source filename so the
  persisted record does not expose a random temporary filename;
- do not buffer the upload or multipart body in memory;
- duplicate content returns `409 import.duplicate` with the existing file hash
  so the UI can navigate to it;
- success returns `201 Created`, the imported file summary, and `Location`.

The initial UI imports one file at a time. Batch upload and progress/resume are
separate follow-up contracts.

The interactive draft, three-area tagging workspace, relationship
visualization, and side-effect-free evaluation workflow are specified in
[gui-import.ai.md](./gui-import.ai.md).

## Error contract

New library endpoints use one structured envelope:

```json
{
  "error": {
    "code": "tag.relationship_invalid",
    "message": "Tag assignment violates collection relationships",
    "details": {
      "missing_demands": [],
      "active_conflicts": []
    }
  }
}
```

Stable categories:

| HTTP | Example code | Meaning |
| --- | --- | --- |
| `400` | `request.invalid`, `search.invalid`, `tag.value_invalid` | malformed client input |
| `404` | `collection.not_found`, `file.not_found`, `tag.not_found` | resource is unavailable |
| `409` | `import.duplicate`, `file.revision_conflict`, `tag.relationship_invalid` | valid request conflicts with current state |
| `409` | `storage.last_copy_confirmation_required` | destructive confirmation is required |
| `413` | `import.too_large` | request exceeds the upload limit |
| `422` | `collection.invalid` | configured resource exists but cannot be used |
| `500` | `application.internal` | unexpected internal failure |
| `503` | `application.not_ready` | configuration snapshot is not ready |

Messages are safe for display but are not parsed by Vue. Expected Core error
types/sentinels are mapped with `errors.Is`/`errors.As`; substring matching is
forbidden. Unknown internal errors are logged with context and return a generic
message without filesystem paths.

Method rejection continues to include `Allow`. API paths never use the Vue
HTML fallback.

## Core gaps to implement before each UI feature

| Capability | Existing support | Required Core/repository work |
| --- | --- | --- |
| status/config snapshot | complete | expose readiness to other route groups safely |
| list/describe collections | catalog is internal | add deterministic collection summaries/schema |
| browse/search | `Core.Search` | bounded `limit + 1` page helper or equivalent |
| file detail | `Core.GetFile` | frontend-safe mapping and opaque revision |
| file content | storage paths are internal | add streamed content-opening workflow |
| import schema | `Core.ImportFields` | expose required/type/value DTOs |
| upload/import | path-based `Core.Import` | explicit source filename and temp-upload adapter |
| import draft evaluation | evaluator exists internally | add side-effect-free canonical draft/evaluation workflow |
| tag completion | `SearchCollectionTags` | bounded DTO mapping |
| allowed values | `AllowedValues` | structured reason DTO mapping |
| tag mutation | `SetTag`/`RemoveTag` | conditional revision check in transaction |
| storage mutation | `AddStorage`/`RemoveStorage` | conditional revision and preflight last-copy confirmation |

Do not bypass Core by reading YAML, SQLite, or storage files directly from
`webapi`.

## Security and operational boundaries

- The standalone server remains loopback-only for this slice. Authentication,
  TLS, and remote-network binding must be specified before exposure to other
  hosts.
- Treat collection, tag, hash, filename, and multipart values as untrusted.
- Never accept a browser-provided filesystem path.
- Do not expose absolute configuration, source, database, or storage paths.
- Bound page size, completion result size, multipart body size, and small JSON
  request bodies.
- Propagate request cancellation through Core, database, copy, hash, and
  content operations.
- Mutations are not automatically retried by Vue.
- Search GETs may be retried after cancellation; upload and mutation requests
  require explicit user action after an uncertain transport failure.

## Implementation order

Each phase is a separately reviewed and committed vertical slice.

### Phase 0: freeze contracts and error mapping

- [ ] Confirm route names, DTOs, page sizes, and upload limit.
- [ ] Define structured API errors and Core error classification.
- [ ] Add contract tests before route implementation.

### Phase 1: collection selection

- [ ] Add Core collection list/detail workflows without persistent sessions.
- [ ] Add collection endpoints and DTO tests.
- [ ] Add Vue Router, application shell, chooser, and URL selection.
- [ ] Redirect the ready default collection from `/`.

Acceptance: a ready user can select a valid collection and refresh/bookmark
the selected route without leaving an open database session.

### Phase 2: read-only search and file metadata

- [ ] Add bounded search page responses and cancellation tests.
- [ ] Add file summary/detail mapping and stable ETags.
- [ ] Add tag completion endpoints.
- [ ] Build search input, file grid, pagination, and file metadata view.

Acceptance: browsing, tag search, empty states, and navigation work identically
in server and Wails modes.

### Phase 3: file content

- [ ] Add Core content-opening workflow with deterministic storage fallback.
- [ ] Implement streamed GET/HEAD and single-range handling.
- [ ] Add safe preview selection and download fallback in Vue.

Acceptance: large content is streamed and browser-supported media can be
viewed without exposing storage paths.

### Phase 4: one-file import

- [ ] Implement the interaction and responsive behavior in
  [gui-import.ai.md](./gui-import.ai.md).
- [ ] Add side-effect-free Core and HTTP import-draft evaluation.
- [ ] Extend Core import metadata for the original display filename.
- [ ] Add import schema and streaming multipart endpoint with limits/cleanup.
- [ ] Map duplicate and validation errors.
- [ ] Build typed import form and success navigation.

Acceptance: browser and Wails can import the same file through the same API,
and no temporary file remains after success, error, or cancellation.

### Phase 5: file tag editing

- [ ] Add repository/Core conditional revision mutations.
- [ ] Add tag set/remove and allowed-values endpoints.
- [ ] Map relationship evaluation into structured details.
- [ ] Build typed editors, hints, stale-edit recovery, and mutation tests.

Acceptance: valid edits persist; invalid or stale edits are actionable and do
not partially mutate the file.

### Phase 6: storage editing

- [ ] Add revision-safe storage mutations and last-copy preflight.
- [ ] Add explicit destructive confirmation protocol.
- [ ] Build storage controls and physical/logical result messaging.

Acceptance: storage copies are manageable without disguising physical deletion
as an ordinary tag edit.

### Phase 7: end-to-end release gate

- [ ] Test all UI states and cancellation behavior.
- [ ] Test both runtime modes against one API contract.
- [ ] Test upload limits, duplicate import, stale edits, and last-copy removal.
- [ ] Run `go tool task check`, race tests, server build, and Wails build.
- [ ] Update human GUI/HTTP documentation and navigation.

## Deferred contracts

- batch, directory, drag-and-drop queue, resumable, and progress uploads;
- native Wails file picker and trusted path import;
- thumbnails, transcoding, and background jobs;
- editing tag/relationship/collection/storage YAML configuration;
- deleting a file independently of storage removal;
- source history/path disclosure controls;
- saved searches, sorting choices, facets, and total counts;
- authentication, authorization, TLS, and non-loopback binding;
- remote/rclone storage providers and background synchronization.
