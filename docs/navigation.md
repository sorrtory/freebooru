# FreeBooru project map

Update this file when files are added, removed, or given a different
responsibility.

## Dependency direction

Executables are thin frontends. They call Core instead of implementing config,
filesystem, or database rules.

```text
cmd/freebooru-cli ─┐
cmd/freebooru-gui ─┼─> internal/bootstrap ─> internal/core
cmd/freebooru-server┘                         │
                                              ├─> internal/config
                                              └─> internal/collection
```

- `bootstrap` constructs Core with concrete dependencies.
- `core` contains workflows shared by all frontends.
- `config` owns YAML configuration and validation.
- `collection` owns SQLite connections and migrations.

## Project tree

```text
freebooru/
├── AGENTS.md                         Repository rules for contributors and agents
├── README.md                         Project introduction
├── Taskfile.yml                      Format, lint, test, fix, and CLI tasks
├── go.mod                            Go version and dependencies
├── go.sum                            Dependency checksums
├── .golangci.yml                     Linter and formatter configuration
├── .gitignore                        Ignored files
├── .vscode/
│   └── settings.json                 Repository editor settings
│
├── cmd/
│   ├── freebooru-cli/
│   │   ├── main.go                   CLI process entry point and exit handling
│   │   ├── root.go                   Cobra root, --verbose, logger, command registration
│   │   ├── init.go                   `freebooru-cli init`
│   │   ├── init_test.go              Initialization command tests
│   │   ├── config.go                 `config check` and diagnostic rendering
│   │   └── config_test.go            Diagnostic rendering tests
│   ├── freebooru-server/
│   │   └── main.go                   Server startup; HTTP server is not implemented
│   └── freebooru-gui/
│       └── main.go                   GUI placeholder
│
├── internal/
│   ├── bootstrap/
│   │   ├── bootstrap.go              Resolve defaults and construct Core
│   │   └── bootstrap_test.go         Bootstrap integration tests
│   │
│   ├── core/
│   │   ├── core.go                   Shared load, check, and init workflows
│   │   └── core_test.go              Workflow and database lifecycle tests
│   │
│   ├── collection/
│   │   ├── database.go               SQLite open, connection settings, ping, and close
│   │   ├── database_test.go          Real SQLite lifecycle tests
│   │   ├── schema.go                 Schema initialization and version checks
│   │   └── migrations/
│   │       └── 001_initial.sql        Initial schema-migrations table
│   │
│   └── config/
│       ├── app.go                    Application config, defaults, loading, validation
│       ├── storage.go                Storage config and validation
│       ├── collections.go            Collection config and tag references
│       ├── tags.go                   Tag, predefined-value, and relationship models
│       ├── tag_validate.go           Tag-local schema validation
│       ├── tag_value.go              Typed tag-value validation
│       ├── tag_value_test.go         Tag type and boundary tests
│       ├── graph.go                  Directed relationship graph types
│       ├── graph_build.go            Deterministic forward/reverse graph indexes
│       ├── graph_check.go            Cross-file relationship validation
│       ├── graph_compile.go          Compile valid edges into a usable graph
│       ├── graph_contradiction.go    Demand/conflict semantic comparison
│       ├── graph_query.go            Relationship and backlink queries
│       ├── graph_validate.go         Complete graph validation pipeline
│       ├── predicate_compile.go      Predicate shape and combination dispatcher
│       ├── predicate_is.go           Typed scalar equality compilation
│       ├── predicate_set.go          Predefined-value membership compilation
│       ├── predicate_bounds.go       Numeric, temporal, and regex compilation
│       ├── paths.go                  Standard FreeBooru config paths
│       ├── path.go                   $HOME/~ expansion and absolute-path validation
│       ├── path_test.go              Path expansion tests
│       ├── name.go                   Shared identifier validation
│       ├── yaml.go                   Generic strict YAML reader and writer
│       ├── init.go                   Non-destructive default provisioning
│       ├── init_test.go              Provisioning and conflict tests
│       ├── discover.go               Recursive YAML file discovery
│       ├── decode.go                 Strict YAML document decoding
│       ├── validate.go               Domain validation orchestration
│       ├── catalog.go                Immutable catalog types and lookup API
│       ├── catalog_builtin.go        Reserved tags derived from domain config
│       ├── catalog_build.go          Catalog loading, normalization, and indexing
│       ├── catalog_clone.go          Defensive copies for immutable query results
│       ├── catalog_group.go          Implicit group index construction
│       ├── catalog_query.go          Catalog lookup and prefix-search methods
│       ├── catalog_resolve.go        Collection reference resolution
│       ├── catalog_test.go           Catalog lookup, duplicate, and default tests
│       ├── diagnostic.go             Structured configuration diagnostics
│       ├── diagnostic_test.go        Diagnostic and document-rule tests
│       └── config_test.go            Config round-trip and validation tests
│
└── docs/
    ├── navigation.md                 This file map
    ├── cli.md                        Human CLI reference
    ├── config-example.md             Human configuration examples
    ├── config-spec.md                Human configuration field reference
    ├── config-best_practise.md       Naming and configuration conventions
    ├── dependencies.md               Selected libraries and tools
    ├── config.ai.md                  Implementation-facing config contract
    ├── mvp.ai.md                     MVP goal and completion criteria
    ├── todo.ai.md                    Ordered implementation checklist
    └── user-story.excalidraw         Editable user-flow diagram
```
