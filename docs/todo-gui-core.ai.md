# FreeBooru GUI Core status TODO

This document defines the next GUI vertical slice after
[todo-gui.ai.md](./todo-gui.ai.md): connect the shared web and Wails frontend
to a real `core.Core` instance and expose configuration readiness without
opening a collection.

The slice is deliberately read-only. Collection lifecycle, search, media,
imports, and tag mutations remain deferred until both runtimes can present the
same actionable application status.

## Goal

Both the standalone server and Wails must construct one Core and inject it into
the shared HTTP API. Vue must display whether FreeBooru configuration is ready,
the configured default collection, and every structured diagnostic.

```text
Vue GET /api/v1/status
            |
       shared webapi
            |
        core.Core
       /         \
server process   Wails process
```

Correcting configuration on disk and pressing `Retry` must refresh the status
without restarting either process.

## Fixed decisions

- Keep `GET /api/v1/hello` as a transport-only liveness endpoint.
- Add `GET /api/v1/status` for real application readiness.
- Construct exactly one Core per process and inject it into `webapi`.
- Define the Core dependency as a small consumer-owned interface in `webapi`;
  do not make the adapter depend on unrelated Core workflows.
- Evaluate status on every `/status` request by calling `LoadConfig`, then
  `CheckConfig` only when loading succeeded.
- Propagate `request.Context()` through the complete status workflow.
- `ready` means application YAML loaded and configuration diagnostics contain
  no errors. Warnings do not make the application unready.
- Configuration problems are expected application state, not transport
  failure. Return `200 OK` with `ready: false` and diagnostics.
- Preserve structured diagnostic fields at the HTTP boundary; do not make Vue
  parse Go error strings.
- Convert a `LoadConfig` failure into one stable API diagnostic with code
  `application.config_load`.
- Include the runtime mode in status so the hello screen does not need two
  requests to render its connected state.
- Keep the API model independent from `config.Diagnostic` and other Go domain
  structures. Mapping belongs to `webapi`.
- Keep status stateless in Vue beyond component/composable state. Do not add
  Pinia for this slice.
- Do not automatically create or repair configuration. Provisioning remains
  the explicit responsibility of `freebooru-cli init`.
- Do not open a collection or database while calculating status.
- Do not add authentication or remote-network binding in this slice; the HTTP
  server remains loopback-only.

## HTTP contract

### Request

```http
GET /api/v1/status
Accept: application/json
```

### Ready response

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{
  "ready": true,
  "mode": "server",
  "default_collection": "main",
  "diagnostics": []
}
```

### Configuration-error response

```json
{
  "ready": false,
  "mode": "desktop",
  "default_collection": "main",
  "diagnostics": [
    {
      "severity": "error",
      "code": "reference.missing",
      "message": "tag reference is missing",
      "file": "/home/user/.config/freebooru/collections/main.yaml",
      "document": 1,
      "field": "tags.require[0]"
    }
  ]
}
```

When application YAML cannot load, `default_collection` is an empty string and
the response contains:

```json
{
  "severity": "error",
  "code": "application.config_load",
  "message": "load application configuration: ...",
  "file": "",
  "document": 0,
  "field": ""
}
```

The message must remain actionable, but tests must assert its stable context
rather than the complete OS-dependent error text.

### Method and route behavior

- Other methods on `/api/v1/status` return `405 Method Not Allowed`.
- The response includes `Allow: GET` for method rejection.
- API errors use the existing JSON error envelope.
- `/api/v1/status` never returns the Vue HTML fallback.

## Target API seam

The exact names may be refined during implementation, but the dependency must
remain no broader than:

```go
type Application interface {
    LoadConfig(context.Context) error
    CheckConfig(context.Context) config.Diagnostics
    AppConfig() config.AppConfig
}

func New(mode Mode, app Application) (http.Handler, error)
```

Constructor validation must reject an invalid mode and a nil application. A
typed-nil interface must not be accepted accidentally. If typed-nil validation
would require reflection, prefer a concrete, non-nil adapter function over a
reflection-heavy generic check.

The handler must call the injected interface only for `/status`; `/hello`
remains available even when status is unready.

## Process startup behavior

### Server

The server needs a port before it can answer `/status`:

1. Construct Core. Failure is fatal.
2. Attempt `LoadConfig` once to obtain `http_port`.
3. When it succeeds, use the configured port.
4. When it fails, log one warning and use
   `config.DefaultAppConfig().HTTPPort`.
5. Start the loopback server and let `/status` repeat the load/check workflow.

The preflight error is logged at the process boundary and is not retained as
mutable global state. A later successful `/status` request must report ready
without restarting the server.

### Wails

1. Create the process logger.
2. Construct Core through `bootstrap.NewCore`.
3. Inject Core into `webapi`.
4. Start Wails even when configuration is missing or invalid.
5. Let the first Vue `/status` request determine readiness.

Core construction or API construction failure is fatal. Configuration state is
not fatal.

## Phase 0: finalize documentation and baseline

- [ ] Extend [http-server.md](./http-server.md) with the complete status
  request, response, method, and readiness contract.
- [ ] Extend [gui.md](./gui.md) with the configuration-ready, warning, and
  error screen behavior.
- [ ] Confirm `GET /api/v1/hello` remains supported and documented.
- [ ] Record that invalid configuration no longer prevents the server UI from
  starting on the default port.
- [ ] Run `go tool task check` before implementation and record unrelated
  failures.

Acceptance:

- HTTP semantics and process startup behavior are defined before code changes.

## Phase 1: define status DTOs and Core seam

- [ ] Add API-only `StatusResponse` and `DiagnosticResponse` types.
- [ ] Use explicit JSON tags for every response field.
- [ ] Restrict `mode` and diagnostic `severity` to documented string values.
- [ ] Add the minimal `Application` interface required by status inspection.
- [ ] Change `webapi.New` to accept and validate the application dependency.
- [ ] Update existing hello tests and both executable call sites for the new
  constructor.
- [ ] Do not export internal handler implementation types unnecessarily.

Acceptance:

- Web API construction fails early for missing required dependencies.
- Hello behavior remains byte-compatible except where JSON key order was never
  contractual.

## Phase 2: implement status inspection

- [ ] Register only `GET /api/v1/status` on the new route.
- [ ] Pass `request.Context()` to `LoadConfig` and `CheckConfig`.
- [ ] Skip `CheckConfig` and `AppConfig` when `LoadConfig` fails.
- [ ] Map a load failure to `application.config_load` and `ready: false`.
- [ ] Map every config diagnostic without dropping severity, code, message,
  file, document, or field.
- [ ] Set `ready` from `Diagnostics.HasErrors()` after a successful load.
- [ ] Include `default_collection` after a successful application load,
  including when catalog diagnostics make status unready.
- [ ] Preserve deterministic diagnostic ordering from Core.
- [ ] Return an empty JSON array, not `null`, when there are no diagnostics.
- [ ] Reject other methods with JSON `405` and `Allow: GET`.

Acceptance:

- Repeating the request re-reads configuration and reflects corrected files.
- Status inspection never opens the collection database.

## Phase 3: test the API contract

- [ ] Use a small fake implementing only the `Application` interface.
- [ ] Test ready status with the default collection.
- [ ] Test warning-only status remains ready.
- [ ] Test one and multiple blocking diagnostics.
- [ ] Test all diagnostic fields survive JSON mapping.
- [ ] Test application load failure and stable diagnostic code.
- [ ] Prove `CheckConfig` is not called after load failure.
- [ ] Test empty diagnostics encode as `[]`.
- [ ] Test runtime mode for server and desktop.
- [ ] Test method rejection and unknown API paths.
- [ ] Test request cancellation reaches the fake application.
- [ ] Keep these as fast `httptest.ResponseRecorder` unit tests; no TCP socket
  is needed for the handler contract.

Acceptance:

- Tests constrain observable JSON and call ordering without depending on Core
  internals.

## Phase 4: wire Core into the server

- [ ] Reuse the Core already created in `freebooru-server`.
- [ ] Separate the port preflight from readiness response construction.
- [ ] Use configured `http_port` after a successful application load.
- [ ] Use the documented default port after load failure.
- [ ] Log the preflight failure once with structured `slog` attributes.
- [ ] Pass Core into `webapi.New`.
- [ ] Keep configuration diagnostics non-fatal.
- [ ] Preserve graceful shutdown and existing server timeouts.
- [ ] Add focused tests for configured-port and fallback-port selection without
  binding a real port.

Acceptance:

- Missing or invalid application YAML still allows the Vue application and
  status endpoint to start at the default loopback address.

## Phase 5: wire Core into Wails

- [ ] Create a structured logger at the Wails process boundary.
- [ ] Construct Core once through `bootstrap.NewCore` before `wails.Run`.
- [ ] Pass the same Core instance into `webapi.New`.
- [ ] Do not call `LoadConfig` as a fatal Wails startup requirement.
- [ ] Preserve the existing API middleware and no-localhost-port architecture.
- [ ] Update the middleware test to use a fake application and assert the real
  desktop status response.

Acceptance:

- Wails starts with missing configuration and serves a desktop-mode unready
  status through its internal handler.

## Phase 6: implement the Vue status client

- [ ] Add typed frontend `ApplicationStatus` and `Diagnostic` models matching
  the HTTP schema.
- [ ] Add `getStatus` using relative `fetch("/api/v1/status")`.
- [ ] Validate critical runtime fields at the boundary before returning data
  to components.
- [ ] Treat non-2xx responses, malformed JSON, invalid mode, and invalid
  readiness fields as connection failures.
- [ ] Preserve the complete diagnostic list and stable source metadata.
- [ ] Unit-test ready, unready, malformed, and HTTP-error responses.

Acceptance:

- Vue does not import Go/Wails models and does not parse diagnostic messages.

## Phase 7: refactor the hello screen into focused UI

- [ ] Replace the hello request with the status request.
- [ ] Move request state and retry behavior into a `useApplicationStatus`
  composable once `App.vue` would otherwise mix orchestration and presentation.
- [ ] Add a focused status summary component if diagnostics make `App.vue`
  responsible for multiple independent UI sections.
- [ ] Preserve explicit loading, connection-error, and retry states.
- [ ] Ready state shows runtime transport and default collection.
- [ ] Warning state remains operational and displays warning count/details.
- [ ] Unready state clearly says configuration needs attention.
- [ ] Render diagnostic code, message, severity, file, document, and field when
  present.
- [ ] Do not expose empty source fields as visual noise.
- [ ] Keep Retry at least 44 by 44 CSS pixels and keyboard accessible.
- [ ] Preserve mobile-first behavior at 375, 768, 1024, and 1440 CSS pixels.
- [ ] Use semantic status/alert markup without announcing the entire diagnostic
  list repeatedly during refresh.

Acceptance:

- The same screen handles ready, warnings, configuration errors, connection
  errors, and recovery without a full page reload.

## Phase 8: test frontend behavior

- [ ] Test loading to ready state.
- [ ] Test server and desktop transport labels.
- [ ] Test default collection rendering.
- [ ] Test warning-only status remains presented as usable.
- [ ] Test blocking diagnostics render as unready.
- [ ] Test optional diagnostic source fields are omitted when empty.
- [ ] Test connection failure and retry recovery.
- [ ] Test configuration repair and retry recovery using two status responses.
- [ ] Keep tests focused on accessible text and roles rather than CSS classes.

Acceptance:

- UI tests describe user-observable behavior and do not depend on component
  implementation details.

## Phase 9: end-to-end verification

- [ ] Run frontend tests and production build.
- [ ] Run all Go handler and executable tests.
- [ ] Run `go tool task check`.
- [ ] Build the standalone server and Wails application.
- [ ] With valid configuration, verify both modes report `ready: true` and the
  same default collection.
- [ ] With invalid configuration, verify both modes remain accessible and
  report matching structured diagnostics.
- [ ] Correct the invalid file and verify Retry recovers without restarting.
- [ ] Verify Wails still reports `desktop` through the Vite development proxy.
- [ ] Verify the committed production bundle matches the current Vue source.
- [ ] Update [navigation.md](./navigation.md) for files added or moved.

Final acceptance:

- Both processes own one real Core and expose it through the same narrow API.
- Configuration failure is visible and recoverable instead of preventing UI
  startup.
- Ready semantics and diagnostic JSON are stable and tested.
- No collection or database is opened by status inspection.
- No search, import, media, tag mutation, authentication, or remote binding was
  pulled into this slice.

## Deferred next slices

After this checklist is complete, implement separately and in order:

1. list, select, open, and close collections;
2. read-only paginated search;
3. file metadata and media-content delivery;
4. multipart browser import with explicit limits;
5. native Wails file selection and path-based import;
6. tag editing and relationship diagnostics;
7. authentication, TLS, and configurable remote-network binding.
