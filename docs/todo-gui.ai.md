# FreeBooru hello GUI implementation TODO

This document defines the smallest vertical slice that proves one Vue 3
application can run both as a remotely served web application and as an
embedded Wails desktop application. It is intentionally narrower than the
complete FreeBooru GUI.

The human-facing platform notes belong in [gui.md](./gui.md). The HTTP server
contract belongs in [http-server.md](./http-server.md). Update those documents
before expanding this checklist beyond the hello slice.

## Goal

The same Vue production bundle must display `Hello FreeBooru` after calling:

```http
GET /api/v1/hello
```

The request must reach the same Go `http.Handler` in both environments:

```text
Vue fetch("/api/v1/hello")
            |
       shared handler
        /         \
freebooru-server  Wails AssetServer
TCP HTTP server   internal request handling
```

The Wails application must not open a localhost TCP port for this slice.

## Fixed decisions

- Use Wails v2 for the first implementation.
- Use Vue 3, Vite, TypeScript, Composition API, and
  `<script setup lang="ts">`.
- Produce one Vue production bundle and embed that bundle into Go.
- Vue calls relative `/api/v1/...` URLs with ordinary `fetch`.
- Both executables reuse one standard-library `http.Handler` for the API.
- `freebooru-server` serves both the embedded Vue assets and the API.
- `freebooru-gui` serves the assets through Wails and routes `/api/v1/...`
  requests to the shared handler through the Wails AssetServer.
- Keep Wails imports out of Vue components and out of `internal/core`.
- Keep HTTP request decoding and response encoding out of `internal/core`.
- Use explicit manual construction in each `main`; do not add a DI framework.
- Do not add Vue Router, Pinia, authentication, uploads, database access,
  collection opening, or native dialogs to the hello slice.
- Do not add a second Go module for the GUI.

## Target layout

The exact generated filenames may follow the selected Wails template, but the
responsibilities must remain:

```text
freebooru/
├── frontend/                         Vue/Vite source
│   ├── package.json
│   ├── vite.config.ts
│   ├── index.html
│   └── src/
│       ├── main.ts
│       ├── App.vue
│       └── api.ts                    Typed hello request
├── internal/
│   ├── webapi/
│   │   ├── handler.go               Shared /api/v1 router
│   │   └── handler_test.go           httptest contract tests
│   └── webui/
│       ├── assets.go                Embedded production assets
│       └── dist/                    Generated Vite output
├── cmd/
│   ├── freebooru-server/main.go     TCP server and graceful shutdown
│   └── freebooru-gui/main.go        Wails application wiring
└── wails.json                       Wails build configuration
```

If Wails requires the embedded assets to be colocated differently, preserve
one frontend source tree and one generated bundle rather than duplicating the
Vue application.

## Phase 0: record the contract

- [x] Expand [http-server.md](./http-server.md) with the hello endpoint,
  response schema, content type, allowed method, and error behavior.
- [x] Make [gui.md](./gui.md) state that remote web and Wails use the same Vue
  source and the same HTTP-shaped API.
- [x] Correct repository documentation that still describes Tauri as the GUI
  implementation.
- [x] Record Wails v2 as the selected major version in
  [dependencies.md](./dependencies.md).
- [x] Run the existing `go tool task check` and record unrelated baseline
  failures before adding generated files or dependencies.

Acceptance:

- The hello endpoint and the two runtime paths are unambiguous before code is
  added.

## Phase 1: obtain and normalize the Wails Vue scaffold

The upstream scaffold can be inspected with:

```bash
wails init -n myproject -t vue-ts
```

- [x] Generate the scaffold in a temporary directory, or otherwise ensure it
  cannot overwrite the existing repository.
- [x] Copy only the required Wails configuration, build metadata, icons, and
  Vue/Vite files into FreeBooru.
- [x] Rename scaffold identifiers from `myproject` to `freebooru`.
- [x] Integrate Wails into the existing `github.com/sorrtory/freebooru` module;
  do not retain the scaffold's `go.mod`, sample Go application, or generated
  hello bindings.
- [x] Remove template counters, greetings, logos, and sample CSS.
- [x] Keep one package manager and commit its lockfile.
- [x] Ignore dependency and Wails packaging output. Commit the small Vite
  production bundle so ordinary Go builds work directly from a checkout.

Acceptance:

- The frontend type-checks and builds independently.
- The existing Go module and command layout remain authoritative.
- No template demo behavior remains.

## Phase 2: implement the minimal Vue screen

- [x] Add a typed `HelloResponse` with `message` and `mode` fields.
- [x] Add one API function that calls `fetch("/api/v1/hello")`.
- [x] Treat a non-2xx response or invalid JSON as a visible connection error.
- [x] Make `App.vue` own only the hello-screen orchestration for this slice.
- [x] Display a loading state while the request is pending.
- [x] Display `Hello FreeBooru` and the returned mode on success.
- [x] Display an actionable error and retry button on failure.
- [x] Keep the page usable at narrow and desktop window widths.
- [x] Do not detect Wails in frontend code; the relative API URL is identical
  in both environments.

Acceptance:

- A mocked successful request renders the message and mode.
- A failed request renders an error and can be retried.
- The component contains no Wails-generated imports.

## Phase 3: implement the shared hello API

- [x] Add `internal/webapi` with a constructor returning `http.Handler`.
- [x] Register only `GET /api/v1/hello` for this slice.
- [x] Return `Content-Type: application/json` and a stable JSON object:

  ```json
  {"message":"Hello FreeBooru","mode":"server"}
  ```

- [x] Inject the mode explicitly when constructing the handler so the Wails
  executable can return `desktop` without globals or runtime detection.
- [x] Return `405 Method Not Allowed` for unsupported methods.
- [x] Return JSON errors for API failures; never return the Vue fallback page
  for an `/api/` request.
- [x] Use `httptest` to cover the response body, content type, method handling,
  and unknown API route behavior.
- [x] Do not initialize Core for a response that is intentionally constant.

Acceptance:

- The complete API contract is tested without starting a TCP listener or
  Wails.

## Phase 4: embed and serve the Vue production bundle

- [x] Configure Vite to write its production output to the directory embedded
  by `internal/webui`.
- [x] Embed the generated assets with `//go:embed`.
- [x] Expose an `fs.FS` rooted at the bundle so callers do not depend on the
  physical build directory prefix.
- [x] Add an HTTP static handler that serves `index.html` at `/` and hashed
  assets at their generated paths.
- [x] Keep `/api/` routing ahead of any single-page-application fallback.
- [x] Fail the production build clearly when the frontend bundle is absent.

Acceptance:

- A Go test or built server can serve `index.html` and one generated asset from
  the embedded filesystem.
- Unknown `/api/` paths produce an API 404 rather than HTML.

## Phase 5: wire the remote web server

- [x] Make `cmd/freebooru-server` construct the hello handler with mode
  `server`.
- [x] Mount the API handler and embedded Vue assets on one HTTP server.
- [x] Listen on the configured `http_port`; add a narrow override only if the
  existing configuration contract requires one for tests or development.
- [x] Configure explicit read-header, read, write, and idle timeouts.
- [x] Handle process cancellation and shut the server down with a bounded
  context.
- [x] Log startup and terminal errors without logging response bodies or future
  upload contents.
- [x] Add an integration test using `httptest.Server` rather than a fixed port.

Acceptance:

- Opening the server URL loads Vue and displays `Hello FreeBooru` with mode
  `server`.

## Phase 6: wire the Wails desktop application

- [x] Make `cmd/freebooru-gui` construct the same hello handler with mode
  `desktop`.
- [x] Configure the Wails AssetServer with the same embedded Vue assets.
- [x] Route `/api/v1/...` through the shared handler using the Wails
  AssetServer handler or middleware hook.
- [x] Delegate non-API asset requests to Wails' default asset handling.
- [x] Do not bind sample Go methods merely to implement hello.
- [x] Do not start `net/http.Server` or reserve a localhost port.
- [x] Keep platform build tags, including Linux WebKit requirements, in the
  documented Wails commands rather than domain packages.

Acceptance:

- Starting the desktop executable loads the same Vue bundle and displays
  `Hello FreeBooru` with mode `desktop`.
- The desktop application continues to work while the configured HTTP port is
  occupied by another process.

## Phase 7: make development predictable

- [x] Add a Vite development proxy from `/api` to the development
  `freebooru-server` address.
- [x] Document the browser development flow: start the Go server, then Vite.
- [x] Document the Wails development command, including the required Linux
  `webkit2_41` build tag where applicable.
- [x] Decide whether Wails development uses the Vite proxy or the Wails virtual
  handler, and test that exact path; production must still use the virtual
  handler without a TCP port.
- [x] Add Taskfile targets for frontend install, type-check, test, build,
  browser development, Wails development, and production builds.
- [x] Make the repository quality task run frontend checks when frontend
  dependencies are available in the supported development environment.

Acceptance:

- A new contributor has one documented command sequence for browser mode and
  one for desktop mode.
- Frontend errors fail the authoritative project check.

## Phase 8: verify the vertical slice

- [x] Run frontend type-checks, tests, and production build.
- [x] Run Go handler and server integration tests.
- [x] Run `go tool task check` with the repository's documented sandbox cache
  override when required.
- [x] Manually verify the browser build reports mode `server`.
- [x] Manually verify the Wails build reports mode `desktop`.
- [x] Verify both modes issue `GET /api/v1/hello` from the same frontend code.
- [x] Verify the Wails build does not listen on the configured HTTP port.
- [x] Update [navigation.md](./navigation.md) for every file added or moved by
  the implementation.

Final acceptance:

- One Vue source tree and one production bundle serve both runtimes.
- One Go `http.Handler` implements the hello API for both runtimes.
- Remote web uses a normal TCP HTTP server.
- Wails handles the same HTTP-shaped request internally without a TCP port.
- The page has explicit loading, success, failure, and retry behavior.
- No full FreeBooru workflow was pulled into the hello slice prematurely.

## Deferred work

The hello slice must leave these as later, separately specified phases:

- multipart file upload and upload limits;
- native Wails file dialogs and path-based local import;
- authentication, authorization, TLS, and cross-origin deployment;
- collection selection and lifecycle;
- search, thumbnails, media responses, and platform streaming validation;
- import progress reporting and cancellation;
- tag editing and relationship diagnostics;
- Vue Router, shared application state, and offline behavior;
- installers, signing, auto-update, and release packaging.
