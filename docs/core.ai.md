# FreeBooru Core contract

This document defines application-layer invariants shared by the CLI, HTTP
server, and Wails frontend. Core owns workflows and resource lifecycles;
frontends translate their inputs into explicit Core calls.

## Collection identity

Every collection-scoped operation must identify its collection explicitly.

Examples include:

- collection validation and description;
- search and file lookup;
- file content access;
- import and import-field discovery;
- tag lookup, value hints, assignment, replacement, and removal;
- storage assignment and removal.

New collection-scoped Core request types must contain a `Collection string`,
or the method must accept a collection name directly. A frontend must resolve
its default or selected collection before calling Core and pass the canonical
name on every call.

An empty collection name is retained only for compatibility with existing CLI
workflows that define empty as “use the configured default.” New HTTP/Wails
code must not depend on that fallback. Tests for new web API routes must prove
that the route collection is passed through to Core.

Global operations do not take a collection. These include application
initialization, configuration load/check, application status, and listing the
configured collections.

## Session ownership

Core may open a collection database for the duration of one operation and must
close it when that operation returns. Request cancellation must propagate into
the operation and cleanup.

`OpenCollection` and `CloseCollection` provide an explicit retained session
for a single-owner frontend such as a bounded CLI workflow. They are not the
HTTP notion of selecting or opening a collection.

The HTTP API and Wails asset middleware must not retain a Core collection
session across requests. Vue selection is represented by its route, and every
request repeats that collection name. This prevents one browser tab, client,
or collection from changing process-global state for another.

```text
Vue /collections/main
        |
GET /api/v1/collections/main/files
        |
Core.Search(Collection: "main")
        |
request-scoped database lifecycle
```

## Frontend boundaries

- Frontends call Core rather than reading YAML, SQLite, or storage paths.
- Core APIs use frontend-independent request and result types.
- HTTP DTOs are defined and mapped in `webapi`; Core does not expose JSON or
  Wails-specific models.
- Collection names are validated and canonicalized by Core. A frontend may
  normalize user input for navigation, but it is not authoritative.
- Errors retain stable identities or structured types so frontends can map
  them without parsing error strings.

## Concurrency consequence

Passing the collection explicitly prevents ambiguous routing, but it does not
by itself provide parallel collection access. Core currently protects its
configuration snapshot and collection lifecycle with a process-local mutex.
New workflows must preserve correctness and cancellation first; concurrency
may be refined later without changing the explicit collection contract.
