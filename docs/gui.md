# FreeBooru GUI

FreeBooru has one Vue 3 and TypeScript frontend. The production bundle runs in
two environments:

- the remote web application is served by `freebooru-server`;
- the desktop application is embedded in `freebooru-gui` and displayed by
  Wails.

Vue uses relative `/api/v1/...` URLs in both environments. The server handles
those requests over TCP. Wails routes them internally through the same Go
`http.Handler` without opening a localhost port.

## Application readiness

After connecting, Vue requests `GET /api/v1/status`. The status screen has four
observable states:

- loading while Core reads configuration;
- ready with the runtime and default collection;
- ready with warnings and their structured details;
- unready with configuration errors and their structured details.

A connection failure is distinct from invalid configuration. Retry requests
status again, so correcting YAML on disk recovers the screen without restarting
the server or desktop application. The GUI never creates or repairs
configuration automatically; use `freebooru-cli init` for provisioning.

## Collection workspace

The selected collection is the route source of truth. Overview, Files, Tags,
Storage, Upload, and file-detail requests all send that collection name
explicitly to Core. Overview counts link to their corresponding resource page.
The compact header keeps Overview and Tags beside collection selection, uses a
fluid search field with tag/value completion and Popular tags, and leaves Files
and Upload as the right-side daily actions. Storage is reached through Settings.

Tags and Storage present resources imported by the collection before other
configured resources, which can be imported through atomic validated Core
operations. Tag rows expose their groups and link directly into Files search.
Files combines browsing and search: a blank query shows everything,
while its responsive booru-style sidebar searches as the user types and offers
collection hot tags plus keyboard tag/value completion. Upload keeps blocking demands beside
the assignment that caused them and focuses the first actionable issue when an
incomplete import is attempted.

The settings gear opens application-scoped Settings. Appearance remains local
browser state. The safe editable `freebooru.yaml` subset is revision-protected,
atomically persisted, and fully revalidated; filesystem paths are preserved by
Core but are not exposed to the remote frontend.

## Desktop

We use Wails v2 to build the desktop application.

To setup wails on ubuntu, you need to install dependencies:

```bash
sudo apt install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

The initial Vue and Wails files were normalized from the official scaffold:

```bash
wails init -n myproject -t vue-ts
```

Do not run the scaffold command in the repository: it creates a separate Go
module and sample application. Use it only in a temporary directory when
comparing against a newer template.

Install frontend dependencies:

```bash
go tool task frontend:install
```

Run the desktop application on Linux:

```bash
go tool task gui:dev
```

The task supplies the `webkit2_41` build tag required by the documented Ubuntu
packages. During Wails development, Vite proxies `/api` back to Wails' fixed
development server at `127.0.0.1:34115`, so no standalone FreeBooru server is
required. A production desktop build is available as `go tool task gui:build`.

## Web

For browser development, run these commands in separate terminals:

```bash
go tool task server:dev
go tool task frontend:dev
```

Vite proxies `/api` to `http://127.0.0.1:52800`. A production server build is
available as `go tool task server:build`. See [HTTP Server](http-server.md) for
the API and static-file contract.
