# FreeBooru implementation TODO

Implementation-facing execution plan for the MVP defined by
[mvp.ai.md](./mvp.ai.md) and [config.ai.md](./config.ai.md). Complete phases in
order. Keep every phase small enough to review and leave `go tool task check`
passing.

## Fixed decisions

- The only initialization command is `freebooru-cli init`.
- `freebooru-cli config check` remains a read-only validation command.
- Application workflows belong to `core`.
- YAML loading, defaults, path handling, and config validation belong to
  `config`.
- SQLite connection and persistence belong to a new `collection` package.
- Use `database/sql` with `github.com/ncruces/go-sqlite3/driver` and driver name
  `sqlite3`.
- Preserve a connector seam for the driver's Adiantum VFS. Encryption, key
  storage, and key acquisition are not part of the initial connector.
- The config catalog is immutable after a successful build.
- Catalog lookup, relationship traversal, and file-state evaluation are
  separate responsibilities.

## Phase 0: contract and baseline

- [x] Remove `config init` from user-facing command lists in human and AI docs.
- [x] Record `freebooru-cli init` output and failure behavior in CLI docs.
- [x] Run the existing tests before structural changes.
- [x] Record any already-failing checks without changing unrelated code.

Baseline note: `go test ./...` and `go vet ./...` passed. With Go's binary
directory added to `PATH`, `go tool task check` reached `golangci-lint` and
reported the repository's existing `errcheck` and `revive` findings.

Acceptance:

- Documentation names only the top-level initialization command.
- The baseline status is known before implementation starts.

## Phase 1: align the default config model

Update Go structures only as far as needed to represent and round-trip the
documented default application, storage, and collection files.

- [x] Make `AppConfig` represent `lang`, `default_collection`,
  `default_storage_name`, `default_storage_path`, `http_port`, and
  `remove_on_upload`.
- [x] Make `DefaultAppConfig` exactly match `config.ai.md`.
- [x] Keep storage as a list of providers with `name`, `type`, and local `path`.
- [x] Make the default storage list contain the configured default local
  storage.
- [x] Remove obsolete collection `storage` fields.
- [x] Represent collection `tags.require` and `tags.import`.
- [x] Add `storage` to `TagReference`.
- [x] Enforce exactly one reference discriminator: `tag`, `group`, or
  `storage`.
- [x] Allow collection `location` to be omitted and resolve its documented
  absolute default from the collection name.
- [x] Keep `$HOME` and leading `~` expansion in reusable config path helpers.
- [x] Reject relative paths after expansion.
- [x] Add strict YAML round-trip tests for all three default documents.
- [x] Add tests proving generated YAML is accepted by the same loader that
  reads user configuration.

Acceptance:

- Default constructors serialize to the documented shape.
- Serialized defaults load and pass the implemented validation.
- Unknown fields and obsolete fields are rejected rather than ignored.

## Phase 2: non-destructive default provisioning

Replace the current config-only initializer with a reusable provisioning
operation. It creates missing resources but never edits existing YAML.

- [x] Add a config operation that creates required config/data directories.
- [x] Create `freebooru.yaml` first when it is absent.
- [x] Load and strictly verify `freebooru.yaml` before deriving other defaults.
- [x] Create `storage.yaml` only when it is absent.
- [x] Create the default collection YAML only when it is absent.
- [x] Create the tags directory when absent.
- [x] Create the configured default storage directory when absent.
- [x] Create the configured collections data directory when absent.
- [x] Never overwrite, merge, append to, or reformat an existing YAML file.
- [x] If existing `storage.yaml` lacks the configured default storage, return an
  actionable conflict error.
- [x] If an existing default collection conflicts with its derived name or
  location, return an actionable conflict error.
- [x] Make retries safe after partial filesystem creation.
- [x] Wrap every error with the resource and operation that failed.

Provisioning policy:

| Existing state | Behavior |
| --- | --- |
| Resource absent | Create it |
| Valid compatible resource | Preserve it |
| Invalid resource | Fail and report validation errors |
| Valid incompatible resource | Fail and explain the conflict |

Acceptance:

- A new home/config root receives a valid default layout.
- Running provisioning twice leaves file contents unchanged.
- Existing user YAML is never silently modified.
- A failed run can be corrected and retried.

## Phase 3: top-level application initialization

- [x] Add `Core.Init(ctx)` as the application initialization workflow.
- [x] Make it provision and load `freebooru.yaml`.
- [x] Make it provision and verify the default storage and collection.
- [x] Add a top-level Cobra `init` command that only calls `Core.Init` and
  formats the result.
- [x] Delete `newConfigInitCommand` and remove it from the `config` command.
- [x] Keep Cobra handlers free of YAML and filesystem business logic.
- [x] Use `RunE`, `cobra.NoArgs`, `cmd.OutOrStdout`, and `cmd.Context`.
- [x] Add isolated command tests using a fresh command tree and temporary
  configuration roots.

Acceptance:

- `freebooru-cli init` is the only initialization entry point.
- `freebooru-cli config init` is an unknown command.
- Initialization errors propagate without printing Cobra usage noise.

## Phase 4: hello-world SQLite connector

Add `internal/collection` without designing the file/tag persistence schema.

- [x] Add `github.com/ncruces/go-sqlite3` as a direct dependency.
- [x] Register the `github.com/ncruces/go-sqlite3/driver` `database/sql`
  driver.
- [x] Introduce a small connector API: open, initialize, ping, and close.
- [x] Accept an absolute database path and create only its parent directory.
- [x] Open with a `file:` URI suitable for the selected driver.
- [x] Use context-aware database calls.
- [x] Enable foreign-key enforcement for every connection using a driver-safe
  connection option, not a one-time pragma on an arbitrary pooled connection.
- [x] Configure a conservative SQLite connection pool; document the choice.
- [x] Create a minimal versioned schema containing only migration metadata.
- [x] Apply schema initialization transactionally.
- [x] Make schema initialization idempotent.
- [x] Detect a database schema newer than the application supports.
- [x] Close the database if any open/initialize step fails.
- [x] Add integration tests for create, ping, close, reopen, repeated initialize,
  and unsupported schema version.
- [x] Keep encryption configuration out of logs and error strings.

Initial API target:

```go
type Database struct {
    // unexported sql handle
}

func Open(ctx context.Context, path string, options ...Option) (*Database, error)
func (db *Database) Initialize(ctx context.Context) error
func (db *Database) Ping(ctx context.Context) error
func (db *Database) Close() error
```

The option seam may later select an Adiantum VFS and provide key material. Do
not add a placeholder password flag, global key, or environment variable now.

Acceptance:

- The configured main collection database is created and can be reopened.
- The connector uses `ncruces/go-sqlite3`, not another SQLite driver.
- No C toolchain is required by the connector.
- The schema contains no speculative file/tag tables.

## Phase 5: complete application initialization

- [ ] Inject the collection opener into `Core` through manual constructor
  injection.
- [ ] After config provisioning succeeds, have `Core.Init` open the default
  collection database and initialize its schema.
- [ ] Do not retain the database connection after the CLI initialization command
  finishes.
- [ ] Ensure every opened handle is closed on success and failure.
- [ ] Report which config, storage directory, collection config, and database
  were initialized without claiming existing resources were recreated.
- [ ] Test empty initialization, repeated initialization, config conflicts,
  database failure, and recovery after a partial run.

Acceptance:

- One command creates a usable default application layout and SQLite database.
- Repeating the command is safe and non-destructive.
- CLI, server, and GUI can later call the same `Core.Init` workflow.

## Phase 6: config diagnostics and source model

Prepare for a catalog that can support both checking and editing.

- [ ] Replace plain joined errors with structured diagnostics.
- [ ] Include severity, code, message, file, YAML document, and field path when
  available.
- [ ] Distinguish decoding, local validation, duplicate, reference, and graph
  diagnostics.
- [ ] Preserve all discoverable errors instead of returning after the first
  domain error.
- [ ] Keep invalid domain configuration non-fatal to application startup.
- [ ] Keep missing or invalid `freebooru.yaml` fatal to normal operation.
- [ ] Give diagnostics stable codes suitable for GUI filtering and tests.
- [ ] Keep rendering outside `config`; CLI and GUI choose their presentation.

Acceptance:

- `config check` can render multiple errors from different files.
- Config-editor consumers can inspect diagnostic locations without parsing
  human error strings.

## Phase 7: immutable config catalog

Replace loose storage/tag/collection checks with one staged catalog build.

Build stages:

```text
discover -> decode -> validate document -> normalize -> index
         -> resolve references -> validate catalog -> publish snapshot
```

- [ ] Recursively discover `.yaml` and `.yml` files according to the contract.
- [ ] Do not follow symlinked directories.
- [ ] Decode one document per collection file and multiple documents per tag
  file.
- [ ] Preserve source locations on definitions and values.
- [ ] Normalize case-insensitive identifiers once at the indexing boundary.
- [ ] Preserve original spelling for display and diagnostics.
- [ ] Detect duplicate normalized storage, collection, tag, and value names.
- [ ] Invalidate every duplicate definition; never pick a winner by discovery
  order.
- [ ] Build storage, collection, tag, predefined-value, and group indexes.
- [ ] Resolve collection imports and requirements.
- [ ] Make `require` imply `import`.
- [ ] Require at least one storage per collection.
- [ ] Publish an immutable catalog plus diagnostics.
- [ ] Preserve valid independent definitions when other definitions are broken.

Catalog query target:

```go
Tag(name)
Group(name)
Collection(name)
Storage(name)
TagsInGroup(name)
DeclaredValues(tag)
SearchTags(prefix)
```

Search belongs to catalog indexes, not the relationship graph. Start with
normalized prefix matching; fuzzy ranking can be added behind the same API.

Acceptance:

- Lookups are deterministic and case-insensitive.
- The catalog can power command completion without reparsing YAML.
- An invalid collection does not hide an unrelated valid collection.

## Phase 8: full tags and groups

- [ ] Implement `bool`, `text`, `int`, `date`, `datetime`, `value`, and
  `multivalue` definitions.
- [ ] Enforce `0..9223372036854775807` for integers.
- [ ] Strictly parse `YYYY-MM-DD` dates.
- [ ] Strictly parse RFC 3339 datetimes.
- [ ] Enforce unique normalized predefined values.
- [ ] Model `groups: [...]` as implicit many-to-many navigation indexes.
- [ ] Allow tag/group name collisions because references are explicitly typed.
- [ ] Reserve the built-in `storage` tag name.
- [ ] Expose storage providers as values of the built-in multivalue tag.
- [ ] Add table-driven tests for every type and boundary.

Acceptance:

- All graph vertices have stable normalized identities and typed values.
- Groups provide navigation only and add no implicit relationship semantics.

## Phase 9: relationship graph representation

Build the graph for validation and future config-editor queries, not merely as
a cycle checker.

- [ ] Represent a source condition as a tag plus optional predefined value.
- [ ] Represent relationship kind as `suggest`, `demand`, or `conflict`.
- [ ] Represent the target as a tag plus a compiled typed predicate.
- [ ] Preserve `reason` and YAML source location on every edge.
- [ ] Build forward and reverse adjacency indexes.
- [ ] Never infer reverse edges.
- [ ] Keep graph construction deterministic.
- [ ] Provide outgoing and incoming/backlink queries.
- [ ] Filter edges by relationship kind without rebuilding indexes.
- [ ] Allow future editor APIs to find definitions referenced by an edge.

Graph query target:

```go
Outgoing(source)
Incoming(targetTag)
Suggestions(source)
Demands(source)
Conflicts(source)
Backlinks(tag)
```

Acceptance:

- All backlinks to a tag can be retrieved directly.
- Graph queries return stable results suitable for UI display and tests.
- The graph contains explicit config facts only.

## Phase 10: predicate compilation and graph checking

- [ ] Compile relationship predicates once during catalog construction.
- [ ] Validate target availability in every collection that imports the source.
- [ ] Do not auto-import relationship targets.
- [ ] Implement presence matching when no predicate is provided.
- [ ] Implement typed `is`, including boolean absence for `is: false`.
- [ ] Implement multivalue `has` and value/multivalue `not`.
- [ ] Implement inclusive integer `min` and `max`.
- [ ] Implement strict date/datetime `before` and `after`.
- [ ] Compile text `regex` with Go regexp semantics.
- [ ] Permit only `min`+`max`, `before`+`after`, and `has`+`not` combinations.
- [ ] Reject `is` combined with another predicate.
- [ ] Validate referenced predefined values.
- [ ] Reject invalid ranges and regular expressions.
- [ ] Allow `suggest` cycles and `demand` cycles.
- [ ] Reject self-demand and self-conflict.
- [ ] Reject the same source condition demanding and conflicting with the same
  target condition.
- [ ] Report all graph diagnostics with both source and target context.

Acceptance:

- Every predicate and graph rule in `config.ai.md` has positive and negative
  tests.
- Graph checking does not depend on filesystem discovery order.

## Phase 11: state-dependent evaluator and editor hints

Build a consumer over `Catalog + Graph`; do not add mutable file state to the
graph itself.

- [ ] Define a typed representation of the tags currently assigned to a file.
- [ ] Evaluate active source conditions.
- [ ] Return missing demands and active conflicts.
- [ ] Return non-binding suggestions separately from validity errors.
- [ ] Return all declared values through the catalog.
- [ ] Return currently allowed values through the evaluator.
- [ ] Explain why a value is unavailable using the originating edges and
  reasons.
- [ ] Support prefix tag search through the catalog.
- [ ] Support incoming relationship/backlink display through the graph.
- [ ] Keep APIs usable by CLI, GUI, and server without Cobra or HTTP types.

Evaluator query target:

```go
ValidateFile(state)
MissingDemands(state)
ActiveConflicts(state)
Suggestions(state)
AllowedValues(tag, state)
```

Acceptance:

- Validation and config-editor hints use the same compiled rules.
- The editor cannot suggest a value the validator would immediately reject.

## Phase 12: integration and completion

- [ ] Make `Core.CheckConfig` build the catalog and graph and return structured
  diagnostics.
- [ ] Make collection opening require a valid collection and valid dependencies.
- [ ] Keep unrelated valid collections usable.
- [ ] Add CLI rendering for diagnostics and hints.
- [ ] Add shell completion backed by catalog search only when core lookup is
  cheap and side-effect free.
- [ ] Add integration fixtures containing multiple tag files, groups,
  collections, storages, relationships, and broken references.
- [ ] Add regression tests for every documented failure policy.
- [ ] Update status columns in human docs as features become implemented.
- [ ] Run `go tool task check` after every phase and at MVP completion.

## Deferred beyond this TODO

- Enabling Adiantum encryption and defining key management UX.
- Remote/rclone storage.
- Collection tag overrides.
- Relationship targets by group.
- Wildcard imports and exclusions.
- Multiple open collections in one process.
- Persistent GUI window state and live config reload.
