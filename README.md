# FreeBooru

FreeBooru is a local-first file collection manager with strict, typed booru
tags. The MVP provides a Go CLI, YAML configuration, content-addressed local
storage, and one SQLite database per collection.

## Current MVP

- initialize a default collection and storage;
- validate storage, collection, tag, and relationship configuration;
- import regular files transactionally into one or more local storages;
- assign and remove boolean, text, integer, date, datetime, predefined, and
  multivalue tags;
- search persisted files with typed AND queries;
- use the same SQLite collection from independent CLI or future GUI processes.

Remote storage, encryption, and the desktop GUI are not part of this MVP.

## Run from source

FreeBooru requires Go 1.25.10 or newer. From this directory:

```bash
go run ./cmd/freebooru-cli init
go run ./cmd/freebooru-cli config check
go run ./cmd/freebooru-cli import ./example.png
```

For repeated use, build the executable:

```bash
go build -o freebooru-cli ./cmd/freebooru-cli
./freebooru-cli --help
```

Configuration is created under `$XDG_CONFIG_HOME/freebooru`, or
`$HOME/.config/freebooru` when `XDG_CONFIG_HOME` is unset. Default collection
data is stored under `$HOME/.local/share/freebooru`.

See [CLI usage](docs/cli.md), [configuration](docs/config-spec.md), and the
[MVP contract](docs/mvp.ai.md).

## Development

The required quality gate is:

```bash
env PATH="$(go env GOPATH)/bin:$PATH" go tool task check
```

The full MVP release gate adds race, clean-home workflow, and module integrity
checks:

```bash
env PATH="$(go env GOPATH)/bin:$PATH" go tool task release
```

See the [project map](docs/navigation.md) and
[release evidence](docs/release.ai.md).
