# FreeBooru

FreeBooru is a booru-style tagging software with support for multiple file storage backends (praise the rclone).
Its Vue frontend runs as both a Wails desktop application and a web application
served by the FreeBooru HTTP server.

## What does it do?

FreeBooru allows you to organize and tag your files (images, videos, documents, etc.) in a [booru-style](https://safebooru.org/index.php?page=post&s=list) manner.

Also to be DRY with the tags, you have a tag config that allows you to define exact tag types, tag dependencies (and more) allowed within your booru.

`rclone` is used to support a wide variety of cloud storage backends, allowing you to store your files wherever you want. However, FreeBooru can also be used with local storage only or a mix of local and remote. You can easily control the physical location of each file by assigning special tags, which are described in `storage.yaml`

### Philosophy

The FreeBooru is explicit. It doesn't imply anything. It is up to you to define the tags and their relationships strictly otherwise freebooru exits with an error. This is to ensure that your booru is always in a consistent state and you don't have to deal with any unexpected behavior.

### FreeBooru dictionary

- **collection** - one booru instance, one sqlite database, one collection of files and
  bunch of tags, that are imported from the tags folder. You can have multiple collections, each with its own config and storage backends.
- **tag** - a tag is a label that can be assigned to a file. Tag has a type
- **freebooru-cli** - the command line interface for FreeBooru. It is used to manage config validity, collections and tags. It allows you to tag files - save the tags to the database.
- **freebooru-server** - the HTTP server for FreeBooru. It is used to serve the frontend and provide an API for the frontend to interact with the backend.
- **freebooru-gui** - the Wails desktop application for FreeBooru. It is used to provide a GUI for the FreeBooru Core.

## Configuration

The config is done in yaml. It is the key to the minimal tag quantity and quality you want to have in your booru. The config is usually stored in `$HOME/.config/freebooru/` folder.

- **freebooru.yaml** - defines the application config, like the default collection, application settings, etc.
- **storage.yaml** - defines the storage backends you can use. Note that storage is the tag too. So you can copy or delete file by assigning or removing the storage tag.
- **collections** - defines the collections you want to have. Each collection has its own config file.
- **tags** - defines the tags you want to have. Each tag has its own config and can be imported into multiple collections.

Note that config values are not case-sensitive.

### Config rules

See [Configuration Example](docs/config-example.md) for details.
See [Configuration Specification](docs/config-spec.md) for details.

## Code

### Quality

We use `golangci-lint` for linting and formatting.
We also use `task` as a task runner.
It is required to run `go tool task check` to verify the code quality.

### Navigation

See [Navigation](docs/navigation.md) for details.
Update this file if you want to add new files or change the structure of the project.

### Dependencies

See [Dependencies](docs/dependencies.md) for details.
Update this file if you want to add new dependencies or change the existing ones.

## FreeBooru Client

FreeBooru can be called in three different ways:

- cli - calls core directly from the command line
- wails desktop app - calls http server and runs the vuejs frontend in a desktop application
- web app - calls http server and runs the vuejs frontend in a web application

All three ways use the same core and the same configuration.
The frontend is just a way to interact with the core.

See [CLI](docs/cli.md) for command line details.
See [GUI](docs/gui.md) for frontend details.
See [HTTP Server](docs/http-server.md) for backend details.

## Contributing

- First of all, [docs](./docs/) should be the source of truth. If you want to change something or add new features, please update the docs first, then the code.
    - if you need to make a contract or a spec, generate a new file with ai. Don't mess up human and machine-readable docs. The human-readable docs should be in `.md` files, the machine-readable docs should be in `.ai.md` files.
- On making new code changes, always run `go tool task check` to verify the code quality.
    - Go-installed tools may not be on an agent shell's `PATH`. Check
      `$(go env GOBIN)` and `$(go env GOPATH)/bin` before reporting a tool as
      missing.
    - In a restricted agent sandbox where the normal user cache is read-only,
      run the check with a command-scoped writable cache:
      `env PATH="$(go env GOPATH)/bin:$PATH" XDG_CACHE_HOME=/tmp/freebooru-agent-cache go tool task check`.
      This cache override is for sandboxed agents only; developers should use
      their normal writable user cache.
- Always ensure docs and code are in sync.
- We aim to have a minimalistic and clean codebase, if you see something that can be improved, make a TODO inside the code or docs.
