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
                                              ├─> internal/collection
                                              └─> internal/storage
```

- `bootstrap` constructs Core with concrete dependencies.
- `core` contains workflows shared by all frontends.
- `config` owns YAML configuration and validation.
- `collection` owns SQLite connections and migrations.
- `storage` owns content hashing, atomic publication, verification, and deletion.

At runtime, an executable asks `bootstrap` for a Core. Core loads
`freebooru.yaml`, builds the validated config catalog and relationship graph,
then opens a collection session that binds its config subset, evaluator,
SQLite repository, and storage backends. CLI, GUI, and server code call Core
workflows only; they do not read YAML, issue SQL, or coordinate file copies.

## Project tree

```text
freebooru/
├── AGENTS.md                         Repository rules for contributors and agents
├── README.md                         Project introduction
├── Taskfile.yml                      Format, lint, test, release, fix, and CLI tasks
├── go.mod                            Go version and dependencies
├── go.sum                            Dependency checksums
├── .golangci.yml                     Linter and formatter configuration
├── .gitignore                        Ignored files
├── frontend/                         Shared Vue 3 and TypeScript application
│   ├── package.json                  Frontend dependencies and commands
│   ├── package-lock.json             Reproducible frontend dependency lock
│   ├── vite.config.ts                Build output, test environment, and API proxy
│   ├── tsconfig.json                 Browser TypeScript configuration
│   ├── tsconfig.node.json            Vite TypeScript configuration
│   ├── index.html                    Vue application document
│   └── src/
│       ├── main.ts                   Vue application entry point
│       ├── App.vue                   Route outlet shared by web and Wails
│       ├── App.test.ts               Status, diagnostics, and retry recovery tests
│       ├── router.ts                 Status and explicit-collection import routes
│       ├── pages/
│       │   ├── StatusPage.vue        Runtime and configuration readiness screen
│       │   ├── ImportPage.vue        Responsive three-area import workspace
│       │   └── ImportPage.test.ts    Canonical draft workspace interaction tests
│       ├── components/
│       │   ├── AssignedTags.vue      Required and canonical assigned tag cards
│       │   ├── ImportGuidance.vue    Demand and recommendation cards
│       │   ├── TagCatalog.vue        Searchable unassigned-tag catalog
│       │   └── TagField.vue          FreeBooru type-aware assignment editor
│       ├── StatusDiagnostics.vue     Configuration diagnostic list
│       ├── useApplicationStatus.ts   Status loading and retry state
│       ├── useImportDraft.ts         Canonical draft evaluation state and cancellation
│       ├── api.ts                    Typed status and import HTTP client
│       ├── api.test.ts               Status and import HTTP contract tests
│       ├── style.css                 Global theme and reset
│       └── vite-env.d.ts             Vite browser declarations
├── .vscode/
│   └── settings.json                 Repository editor settings
│
├── cmd/
│   ├── freebooru-cli/
│   │   ├── main.go                   CLI process entry point and exit handling
│   │   ├── root.go                   Cobra root, --verbose, logger, command registration
│   │   ├── root_test.go              Root and dynamic command help tests
│   │   ├── app.go                    Shared checked Core construction for data commands
│   │   ├── completion.go             Core-backed tag and value completion
│   │   ├── completion_test.go        Config and persisted-state completion tests
│   │   ├── collection.go             Explicit collection command dispatch
│   │   ├── catalog.go                Collection/tag list and YAML edit commands
│   │   ├── catalog_commands_test.go  List, scope, file filter, and editor tests
│   │   ├── editor.go                 Direct `$EDITOR` execution and config recheck
│   │   ├── files.go                  Collection file/storage assignment listing
│   │   ├── storage.go                Collection storage listing and global editing
│   │   ├── import.go                 Default and explicit collection file import
│   │   ├── import_test.go            Import command routing, output, and validation tests
│   │   ├── tag.go                    Typed tag/storage mutation and lookup commands
│   │   ├── tag_test.go               Tag routing, persistence, output, and validation tests
│   │   ├── search.go                 Typed search and pagination commands
│   │   ├── search_test.go            Default/explicit search and pagination tests
│   │   ├── mvp_test.go               Clean-home persisted CLI release workflow
│   │   ├── init.go                   `freebooru-cli init`
│   │   ├── init_test.go              Initialization command tests
│   │   ├── config.go                 `config check` and diagnostic rendering
│   │   └── config_test.go            Diagnostic rendering tests
│   ├── freebooru-server/
│   │   ├── main.go                   HTTP lifecycle and shared Core/web wiring
│   │   └── main_test.go              Configured and fallback HTTP port tests
│   └── freebooru-gui/
│       ├── main.go                   Wails assets and shared Core/API wiring
│       ├── main_test.go              Desktop API and middleware routing test
│       └── wails.json                Wails build and Vue development configuration
│
├── internal/
│   ├── webapi/
│   │   ├── handler.go               Shared hello and Core status JSON API
│   │   └── handler_test.go          API status, reload, and error contract tests
│   ├── webui/
│   │   ├── assets.go                Embedded Vue production bundle
│   │   ├── handler.go               Static asset and SPA fallback handler
│   │   ├── handler_test.go          Asset, fallback, and TCP integration tests
│   │   └── dist/                    Committed Vite production output
│   ├── bootstrap/
│   │   ├── bootstrap.go              Resolve defaults and construct Core
│   │   └── bootstrap_test.go         Bootstrap integration tests
│   │
│   ├── core/
│   │   ├── assignment.go             Resolve CLI-style strings into typed tag assignments
│   │   ├── assignment_test.go        Assignment syntax, typing, and repetition tests
│   │   ├── evaluation_error.go       Structured invalid-state reasons for frontends
│   │   ├── evaluation_error_test.go  Evaluator diagnostic rendering tests
│   │   ├── import_fields.go           Side-effect-free collection import form schema
│   │   ├── import_fields_test.go      Required/optional field schema tests
│   │   ├── core.go                   Shared init, config snapshot, search, and open workflows
│   │   ├── catalog.go                Frontend-neutral catalog lists and source lookup
│   │   ├── file_list.go              Persisted file/storage assignment listing
│   │   ├── file.go                   Validated collection file lookup workflow
│   │   ├── file_test.go              File lookup and database lifecycle tests
│   │   ├── collection_state.go       Persisted state conversion and integrity checks
│   │   ├── collection_state_test.go  Persisted types, requirements, and relationship tests
│   │   ├── core_test.go              Workflow and database lifecycle tests
│   │   ├── import.go                  Import request preparation and validation
│   │   ├── import_execute.go          Physical and SQLite import transaction workflow
│   │   ├── import_execute_test.go     Import persistence and rollback tests
│   │   ├── import_copies.go           Multi-storage copy and rollback coordination
│   │   ├── import_copies_test.go      Per-copy failure and cancellation cleanup tests
│   │   ├── import_record.go           Typed import-to-database conversion
│   │   ├── import_source.go           Verified remove-on-upload handling
│   │   ├── import_test.go             Import defaults and validation tests
│   │   ├── tag_assignment.go          Canonical typed tag mutation preparation
│   │   ├── tag_assignment_test.go     Tag preparation and persistence-shape tests
│   │   ├── tag_mutation.go            Validated persisted tag set/remove workflows
│   │   ├── tag_mutation_test.go       Tag mutation and required-state tests
│   │   ├── storage_assignment.go      Local storage resolution and assignment helpers
│   │   ├── storage_mutation.go        Physical/logical storage mutation workflows
│   │   ├── storage_mutation_test.go   Storage ordering, rollback, and deletion tests
│   │   ├── search.go                  Validated collection search workflow and defaults
│   │   ├── search_test.go             Search resolution, pagination, and lifecycle tests
│   │   ├── file_relationship.go       Parent/child relationship workflows
│   │   ├── session.go                 Core-owned validated collection lifecycle
│   │   ├── session_test.go            Session ownership, close, and cancellation tests
│   │   ├── hints.go                   Collection tag and persisted value completion APIs
│   │   ├── hints_test.go              Collection filtering and value availability tests
│   │   └── integration_test.go       Multi-domain config integration fixture
│   │
│   ├── collection/
│   │   ├── database.go               SQLite open, connection settings, ping, and close
│   │   ├── database_test.go          Real SQLite lifecycle tests
│   │   ├── database_concurrency_test.go Independent connections, contention, and retry tests
│   │   ├── repository.go             Persisted file records and repository queries
│   │   ├── repository_iterate.go     Bounded-memory collection state streaming
│   │   ├── repository_iterate_test.go Streaming order and failure tests
│   │   ├── repository_create.go      Atomic initial file persistence
│   │   ├── repository_tag.go         Typed tag assignment mutations
│   │   ├── repository_storage.go     Logical storage assignment mutations
│   │   ├── repository_search.go      Snapshot-consistent paginated file search
│   │   ├── repository_search_test.go Typed matching, ordering, and pagination tests
│   │   ├── repository_relationship.go Parent/child relationship persistence and cycle checks
│   │   ├── repository_relationship_test.go Relationship branching, ordering, and validation tests
│   │   ├── search_sql.go             Parameterized typed search SQL compilation
│   │   ├── search_sql_test.go        SQL parameterization and boundary tests
│   │   ├── repository_transaction.go Private write transaction helpers
│   │   ├── repository_test.go        Repository transaction and lookup tests
│   │   ├── schema.go                 Schema initialization and version checks
│   │   └── migrations/
│   │       ├── 001_initial.sql        Initial schema-migrations table
│   │       ├── 002_collection_state.sql Persisted file, tag, and storage state
│   │       └── 003_file_metadata_relationships.sql MIME, interaction timestamps, and file relationships
│   │
│   ├── evaluator/
│   │   ├── evaluator.go               Relationship evaluation and result API
│   │   ├── evaluator_test.go          Demand, conflict, suggestion, and helper tests
│   │   ├── predicate.go               Compiled predicate matching
│   │   ├── state.go                   Immutable validated file tag state
│   │   ├── values.go                  Allowed-value hints and blocking reasons
│   │   ├── state_test.go              State validation and active-source tests
│   │   └── values_test.go             Candidate-value evaluation tests
│   │
│   ├── search/
│   │   ├── query.go                    Strict frontend-independent AND query parser
│   │   ├── query_test.go               Supported and rejected query syntax tests
│   │   ├── resolve.go                  Collection-aware typed query resolution
│   │   └── resolve_test.go             Type, canonicalization, and availability tests
│   │
│   ├── storage/
│   │   ├── local.go                    Local content paths and streamed inspection
│   │   ├── local_write.go              Atomic local content storage and deletion
│   │   ├── local_test.go               Hash, source-type, and inspection tests
│   │   └── local_write_test.go         Storage, corruption, and cleanup tests
│   │
│   └── config/
│       ├── app.go                    Application config, defaults, loading, validation
│       ├── storage.go                Storage config and validation
│       ├── collections.go            Collection config and tag references
│       ├── tags.go                   Tag, predefined-value, and relationship models
│       ├── system_tags.go            Read-only tag-shaped file metadata definitions
│       ├── tag_validate.go           Tag-local schema validation
│       ├── tag_validate_test.go      Tag schema and relationship YAML tests
│       ├── tag_value.go              Typed tag-value validation
│       ├── tag_value_test.go         Tag type and boundary tests
│       ├── graph.go                  Directed relationship graph types
│       ├── graph_build.go            Deterministic forward/reverse graph indexes
│       ├── graph_check.go            Cross-file relationship validation
│       ├── graph_compile.go          Compile valid edges into a usable graph
│       ├── graph_contradiction.go    Demand/conflict semantic comparison
│       ├── graph_query.go            Relationship and backlink queries
│       ├── graph_validate.go         Complete graph validation pipeline
│       ├── graph_test.go             Graph indexing and defensive-copy tests
│       ├── graph_check_test.go       Collection-aware graph rule tests
│       ├── graph_validate_test.go    End-to-end graph validation tests
│       ├── predicate_compile.go      Predicate shape and combination dispatcher
│       ├── predicate_compile_test.go Typed predicate acceptance and rejection tests
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
│       ├── catalog_query.go          Catalog lookup, listing, and prefix-search methods
│       ├── catalog_resolve.go        Collection reference resolution
│       ├── catalog_test.go           Catalog lookup, duplicate, and default tests
│       ├── diagnostic.go             Structured configuration diagnostics
│       ├── diagnostic_test.go        Diagnostic and document-rule tests
│       └── config_test.go            Config round-trip and validation tests
│
└── docs/
    ├── navigation.md                 This file map
    ├── cli.md                        Human CLI reference
    ├── gui.md                        Human web and Wails setup notes
    ├── gui-import.ai.md              Typed three-area import workspace UX contract
    ├── http-server.md                Human HTTP server reference
    ├── config-example.md             Human configuration examples
    ├── config-spec.md                Human configuration field reference
    ├── config-best_practise.md       Naming and configuration conventions
    ├── collection-db.md              Collection SQLite schema and encryption boundary
    ├── dependencies.md               Selected libraries and tools
    ├── core.ai.md                    Shared Core and explicit collection-call contract
    ├── config.ai.md                  Implementation-facing config contract
    ├── mvp.ai.md                     MVP goal and completion criteria
    ├── release.ai.md                 MVP criteria, evidence, and release command
    ├── todo.ai.md                    Ordered implementation checklist
    ├── todo-gui.ai.md                Hello web and Wails implementation checklist
    ├── todo-gui-core.ai.md           Core readiness GUI implementation checklist
    ├── todo-gui-library.ai.md        Collection, search, content, import, and tag UI contract
    └── user-story.excalidraw         Editable user-flow diagram
```
