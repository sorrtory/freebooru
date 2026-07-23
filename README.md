# FreeBooru

FreeBooru is a [booru-style](https://safebooru.org/index.php?page=post&s=list) tagging software which supports multiple file storage backends local and ~~remote (with the help of [rclone](https://github.com/rclone/rclone))~~.
It is written with `Go` + `Vue.js` and designed to be run as a desktop application via [Wails](https://github.com/wailsapp/wails).

Unlike folder-only organization, FreeBooru keeps metadata in explicit YAML
schemas. Tags can have types, predefined values, dependencies, suggestions,
and conflicts, so invalid combinations are rejected before they reach your
collection. This approach also keeps your metadata consistent, so you can have a single source of truth for all your metadata.

## Features

- Typed tags for booleans, text, integers, dates, predefined values, and lists
- Transactional imports into content-addressed local storage
- Multiple collections backed by independent SQLite databases
- Tag relationships for required, suggested, and conflicting metadata
- Typed search across tags and built-in file metadata
- Vue web interface and Wails desktop application
- Scriptable CLI and per-user HTTP server service
- ~~SQLite encryption~~
- ~~Different storage providers via rclone~~
- ~~Config Editor with completions~~

## Installation

Install the CLI from source:

```bash
go install github.com/sorrtory/freebooru/cmd/freebooru-cli@latest
```

This source installation names the command `freebooru-cli`; release packages
install it as `freebooru`.

Linux releases provide `amd64` and `arm64` Debian packages containing the GUI,
CLI, and server:

```bash
version=0.1.0 arch=amd64
curl -LO "https://github.com/sorrtory/freebooru/releases/download/v${version}/freebooru_${version}_${arch}.deb"
sudo apt install "./freebooru_${version}_${arch}.deb"
```

Use `arch=arm64` on ARM systems.

## Quick start


```bash
freebooru init         # initialize default config
freebooru config check # validate configuration
freebooru-gui          # open the GUI
```

Import and find a file from the command line:

```bash
sha=$(freebooru import ./image.png --tag rating:safe)
freebooru tag "$sha" get
freebooru search rating:safe
```

The packaged server runs as a per-user service with systemd on Linux.
Use the CLI to control it:

```bash
freebooru server start
freebooru server status
freebooru server logs --follow
freebooru server stop
```

The server listens on `0.0.0.0:52800` by default. It currently has no built-in
authentication or TLS; use `http_address: 127.0.0.1` for local-only access.

## Configuration

FreeBooru stores configuration in `$XDG_CONFIG_HOME/freebooru`, or
`$HOME/.config/freebooru` when `XDG_CONFIG_HOME` is unset:

- `freebooru.yaml` — application and HTTP settings
- `storage.yaml` — storage backends
- `collections/` — collection definitions
- `tags/` — typed tag schemas and relationships

See the [configuration example](docs/config-example.md) and
[full specification](docs/config-spec.md).

## Interfaces

All three interfaces use the same Core, configuration, and collection data:

- [`freebooru`](docs/cli.md) operates directly on collections and controls the
  packaged server service.
- [`freebooru-gui`](docs/gui.md) embeds the frontend in a native Linux desktop
  application.
- [`freebooru-server`](docs/http-server.md) serves the frontend and JSON API.

Remote storage and encryption are not implemented yet.

## Development

Install [task](https://taskfile.dev/docs/installation#go-tool) and other [dependencies](docs/dependencies.md):

```bash
go get -tool github.com/go-task/task/v3/cmd/task@latest
```

FreeBooru requires Go 1.25.10, Node.js, and the Linux Wails dependencies listed
in the [GUI documentation](docs/gui.md).

```bash
go tool task frontend:install
env PATH="$(go env GOPATH)/bin:$PATH" go tool task check
```

Useful development commands:

```bash
go tool task gui:dev
go tool task server:dev
go tool task cli
```

See the [project map](docs/navigation.md), [dependency notes](docs/dependencies.md),
and [release process](docs/release.ai.md).

## Contributing

Issues and pull requests are welcome. Update the relevant documentation with
behavior changes and run `go tool task check` before submitting a change.

## License

FreeBooru is licensed under the [GNU AGPL v3](LICENSE).
