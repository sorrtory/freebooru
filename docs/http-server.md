# FreeBooru HTTP server

`freebooru-server` serves the Vue production bundle and the versioned JSON API
from the same address. The default address is `127.0.0.1:52800`, using
`http_port` from `freebooru.yaml`.

The first implementation binds to loopback only. Exposing FreeBooru to a
remote network requires a separately specified authentication and TLS policy.

## Static application

| Request | Response |
| --- | --- |
| `GET /` | Embedded Vue `index.html` |
| `GET /assets/...` | Embedded, hashed Vite asset |
| unknown non-API `GET` | Vue `index.html` fallback |

Paths under `/api/` never fall back to HTML.

## Hello endpoint

```http
GET /api/v1/hello
```

Success response:

```http
HTTP/1.1 200 OK
Content-Type: application/json
```

```json
{"message":"Hello FreeBooru","mode":"server"}
```

`mode` is `server` for the TCP server and `desktop` when the same handler runs
inside Wails. Other response fields have identical semantics in both modes.

Other methods on this path return `405 Method Not Allowed` with a JSON error.
Unknown `/api/` paths return `404 Not Found` with a JSON error.

## Application status

```http
GET /api/v1/status
Accept: application/json
```

The endpoint reloads configuration for every request. It does not open a
collection or database. A ready application returns:

```json
{
  "ready": true,
  "mode": "server",
  "default_collection": "main",
  "diagnostics": []
}
```

`mode` is `server` over TCP and `desktop` inside Wails. Configuration problems
are application state rather than transport failures, so they return
`200 OK`, `ready: false`, and structured diagnostics:

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

If `freebooru.yaml` cannot load, `default_collection` is empty and the response
contains an `application.config_load` error diagnostic. Warnings are returned
but do not make the application unready. Other methods return JSON
`405 Method Not Allowed` with `Allow: GET`.

The standalone server attempts to load `http_port` before listening. If that
load fails, it logs a warning and remains accessible on the default port
`52800`, allowing the web interface to explain the configuration problem.

## Deferred server concerns

Multipart upload limits, authentication, authorization, TLS, cross-origin
deployment, media streaming, and progress reporting are outside the hello
contract and require explicit specifications before implementation.
