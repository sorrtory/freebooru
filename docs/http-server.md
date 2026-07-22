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

## Deferred server concerns

Multipart upload limits, authentication, authorization, TLS, cross-origin
deployment, media streaming, and progress reporting are outside the hello
contract and require explicit specifications before implementation.
