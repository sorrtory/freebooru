# FreeBooru GUI

FreeBooru has one Vue 3 and TypeScript frontend. The production bundle runs in
two environments:

- the remote web application is served by `freebooru-server`;
- the desktop application is embedded in `freebooru-gui` and displayed by
  Wails.

Vue uses relative `/api/v1/...` URLs in both environments. The server handles
those requests over TCP. Wails routes them internally through the same Go
`http.Handler` without opening a localhost port.

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
