# FreeBooru implementation TODO

Implementation-facing execution plan for the MVP defined by
[mvp.ai.md](./mvp.ai.md), [config.ai.md](./config.ai.md), and
[collection-db.md](./collection-db.md). Complete phases in order. Keep every
phase small enough to review and leave `go tool task check` passing.

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
  storage, and key acquisition are post-MVP work.
- The config catalog is immutable after a successful build.
- Catalog lookup, relationship traversal, and file-state evaluation are
  separate responsibilities.
- File content identity is lowercase SHA-256 plus byte size.
- Importing content already present in the selected collection fails without
  mutation and reports the existing file.
- Local storage shards content as
  `<root>/<sha256-prefix-2>/<full-sha256>`.
- SQLite stores logical storage assignments; the local backend derives paths
  from storage configuration and content identity.
- MVP import accepts one regular file and rejects directories and symlinks.
- Import assignments use repeated `--tag 'name[:value]'` flags; only the first
  colon is structural.
- Interactive import prompts for required tags and then optionally for other
  imported tags before sending a complete request to Core.
- Import without an explicit storage uses `default_storage_name` only when it
  is available to the selected collection.
- MVP search is AND-only and has deterministic import-time/hash ordering with
  `--limit 100 --offset 0` defaults.
- Opening a collection streams and validates all persisted assignments against
  the current YAML catalog; incompatible state blocks that collection only.

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

- [x] Inject the collection opener into `Core` through manual constructor
  injection.
- [x] After config provisioning succeeds, have `Core.Init` open the default
  collection database and initialize its schema.
- [x] Do not retain the database connection after the CLI initialization command
  finishes.
- [x] Ensure every opened handle is closed on success and failure.
- [x] Print the documented success message only after configuration, storage,
  collection, and database initialization all succeed.
- [x] Test empty initialization, repeated initialization, config conflicts,
  database failure, and recovery after a partial run.

Acceptance:

- One command creates a usable default application layout and SQLite database.
- Repeating the command is safe and non-destructive.
- CLI, server, and GUI can later call the same `Core.Init` workflow.

## Phase 6: config diagnostics and source model

Prepare for a catalog that can support both checking and editing.

- [x] Replace plain joined errors with structured diagnostics.
- [x] Include severity, code, message, file, YAML document, and field path when
  available.
- [x] Establish stable diagnostic code categories. Decoding and local
  validation are emitted now; duplicate, reference, and graph codes are emitted
  by their catalog and graph phases.
- [x] Preserve all discoverable errors instead of returning after the first
  domain error.
- [x] Keep invalid domain configuration non-fatal to application startup.
- [x] Keep missing or invalid `freebooru.yaml` fatal to normal operation.
- [x] Give diagnostics stable codes suitable for GUI filtering and tests.
- [x] Keep rendering outside `config`; CLI and GUI choose their presentation.

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

- [x] Recursively discover `.yaml` and `.yml` files according to the contract.
- [x] Do not follow symlinked directories.
- [x] Decode one document per collection file and multiple documents per tag
  file.
- [x] Preserve source locations on storage, collection, and tag definitions.
- [x] Normalize case-insensitive identifiers once at the indexing boundary.
- [x] Preserve original spelling for display and diagnostics.
- [x] Detect duplicate normalized storage, collection, tag, and value names.
- [x] Invalidate every duplicate definition; never pick a winner by discovery
  order.
- [x] Build storage, collection, tag, predefined-value, and group indexes.
- [x] Resolve collection imports and requirements.
- [x] Make `require` imply `import`.
- [x] Require at least one existing storage per collection.
- [x] Publish an immutable catalog plus diagnostics.
- [x] Preserve valid independent definitions when other definitions are broken.

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

- [x] Implement config definitions for `bool`, `text`, `int`, `date`,
  `datetime`, `value`, and
  `multivalue` definitions.
- [x] Enforce `0..9223372036854775807` for integers.
- [x] Strictly parse `YYYY-MM-DD` dates.
- [x] Strictly parse RFC 3339 datetimes.
- [x] Enforce unique normalized predefined values within each tag.
- [x] Model `groups: [...]` as implicit many-to-many navigation indexes.
- [x] Allow tag/group name collisions because references are explicitly typed.
- [x] Reserve the built-in `storage` tag name.
- [x] Expose storage providers as values of the built-in multivalue tag.
- [x] Add table-driven tests for every type and boundary.

Acceptance:

- All graph vertices have stable normalized identities and typed values.
- Groups provide navigation only and add no implicit relationship semantics.

## Phase 9: relationship graph representation

Build the graph for validation and future config-editor queries, not merely as
a cycle checker.

- [x] Represent a source condition as a tag plus optional predefined value.
- [x] Represent relationship kind as `suggest`, `demand`, or `conflict`.
- [x] Represent the target as a tag plus a compiled typed predicate.
- [x] Preserve `reason` and YAML source location on every edge.
- [x] Build forward and reverse adjacency indexes.
- [x] Never infer reverse edges.
- [x] Keep graph construction deterministic.
- [x] Provide outgoing and incoming/backlink queries.
- [x] Filter edges by relationship kind without rebuilding indexes.
- [x] Allow future editor APIs to find definitions referenced by an edge.

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

- [x] Compile relationship predicates once from an immutable catalog and raw
  graph before evaluation.
- [x] Validate target availability in every collection that imports the source.
- [x] Do not auto-import relationship targets.
- [x] Implement presence matching when no predicate is provided.
- [x] Implement typed `is`, including boolean absence for `is: false`.
- [x] Implement multivalue `has` and value/multivalue `not`.
- [x] Implement inclusive integer `min` and `max`.
- [x] Implement strict date/datetime `before` and `after`.
- [x] Compile text `regex` with Go regexp semantics.
- [x] Permit only `min`+`max`, `before`+`after`, and `has`+`not` combinations.
- [x] Reject `is` combined with another predicate.
- [x] Validate referenced predefined values.
- [x] Reject invalid ranges and regular expressions.
- [x] Allow `suggest` cycles and `demand` cycles.
- [x] Reject self-demand and self-conflict.
- [x] Reject the same source condition demanding and conflicting with the same
  target condition.
- [x] Report all graph diagnostics with source and target context.

Acceptance:

- Every predicate and graph rule in `config.ai.md` has positive and negative
  tests.
- Graph checking does not depend on filesystem discovery order.

## Phase 11: state-dependent evaluator and editor hints

Build a consumer over `Catalog + Graph`; do not add mutable file state to the
graph itself.

- [x] Define a typed representation of the tags currently assigned to a file.
- [x] Evaluate active source conditions.
- [x] Return missing demands and active conflicts.
- [x] Return non-binding suggestions separately from validity errors.
- [x] Return all declared values through the catalog.
- [x] Return currently allowed values through the evaluator.
- [x] Explain why a value is unavailable using the originating edges and
  reasons.
- [x] Support prefix tag search through the catalog.
- [x] Support incoming relationship/backlink display through the graph.
- [x] Keep APIs usable by CLI, GUI, and server without Cobra or HTTP types.

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

## Phase 12: configuration foundation integration

- [x] Make `Core.CheckConfig` build the catalog and graph and return structured
  diagnostics.
- [x] Make collection opening require a valid collection and valid dependencies.
- [x] Keep unrelated valid collections usable.
- [x] Keep CLI rendering for diagnostics. Hint rendering waits for a command
  that presents evaluator results; no placeholder command is added.
- [x] Expose side-effect-free catalog prefix search through Core for future
  shell completion. No current command accepts tag-name arguments.
- [x] Add integration fixtures containing multiple tag files, groups,
  collections, storages, relationships, and broken references.
- [x] Add regression tests for every documented configuration failure policy.
- [x] Update status columns in human docs as features become implemented.
- [x] Run `go tool task check` after every configuration-foundation phase.

Acceptance:

- Core publishes one consistent catalog and compiled graph snapshot.
- Broken domain definitions remain reportable without hiding independent valid
  collections.
- The configuration, relationship, and in-memory evaluation foundation is
  ready for persistence workflows.

## Phase 13: persistence and CLI contracts

Freeze the remaining behavioral contracts before adding tables or commands.

- [x] Document the canonical lowercase SHA-256 content ID and byte-size
  metadata.
- [x] Document sharded local path derivation; extensions are not part of stored
  identity.
- [x] Define duplicate import as a non-mutating error that identifies the
  existing SHA-256 and known source/storage information.
- [x] Define the SQLite representation for every tag type. Boolean false is
  absence; integers remain signed 64-bit; dates retain canonical strings.
- [x] Define how required collection tags are supplied during import.
- [x] Freeze CLI import tag syntax using the documented `tag:value` separator,
  including free-text quoting and rejection rules.
- [x] Make full SHA-256 the authoritative file selector; defer unique prefixes
  unless explicitly added later.
- [x] Freeze the MVP search grammar, pagination, and deterministic result
  ordering.
- [x] Record database/filesystem transaction and rollback ordering.
- [x] State the crash policy: a crash may leave an unreferenced content copy,
  but must not delete the source or commit a row referring to a missing copy.

Acceptance:

- SQL and Cobra code need not invent user-visible semantics.
- Every completion criterion in `mvp.ai.md` maps to a later phase.

## Phase 14: collection schema migration 2

- [x] Add a new embedded migration; never edit an applied migration.
- [x] Add a `file` table keyed by lowercase SHA-256 with byte size and
  timestamps.
- [x] Add a `file_source` table so imported content retains its original path
  and name without making them part of content identity.
- [x] Add `file_tag` and `file_tag_value` tables representing bool, text, int,
  date, datetime, value, and multivalue without ambiguous coercion.
- [x] Add a `file_storage` table keyed by file and normalized storage name.
- [x] Add foreign keys and uniqueness constraints for duplicate detection and
  assignment identity.
- [x] Add checks for SHA-256 shape, non-negative sizes/integers, and legal typed
  value columns where SQLite can enforce them.
- [x] Add indexes needed for tag/value search and storage lookup.
- [x] Apply migrations in version order in one transaction.
- [x] Preserve newer-schema rejection and idempotent initialization.
- [x] Test schema-1 upgrade, reopen, migration rollback, and foreign keys.

Acceptance:

- SQLite can represent all persisted MVP state without storing GUI state,
  configuration caches, or speculative remote-storage fields.

## Phase 15: collection repositories

Keep SQL and row conversion inside `internal/collection`.

- [x] Define persisted file, source, typed tag, and storage models.
- [x] Add context-aware transaction helpers without exposing `*sql.Tx`.
- [x] Insert a new file by SHA-256 and return a typed duplicate error when its
  content identity already exists.
- [x] Record the initial source observation without making it part of content
  identity or changing an existing record on duplicate import.
- [x] Load one file with every typed tag and storage assignment.
- [x] Stream every persisted assignment for collection-open integrity checks
  without loading the complete collection into memory.
- [x] Add, replace, and remove typed tag values transactionally.
- [x] Add and remove logical storage assignments transactionally.
- [x] Delete a file through foreign-key cascades only when its final storage
  assignment is removed.
- [x] Convert persisted assignments into evaluator input without losing types.
- [x] Reject persisted values incompatible with the current catalog.
- [x] Test real SQLite round trips, concurrent readers, and serialized writers.

Acceptance:

- SQLite round-trips every supported `FileState` value.
- Callers do not construct SQL or depend on table layouts.

## Phase 16: local content-addressed storage

- [x] Add a local backend rooted at an absolute path expanded by the config
  boundary.
- [x] Derive `<root>/<sha[0:2]>/<sha>` from a validated lowercase SHA-256.
- [x] Reject malformed hashes, symlinked sources, directories, and non-regular
  files.
- [x] Stream bytes while calculating SHA-256 and size; never read whole files
  into memory.
- [ ] Stage writes in the destination filesystem and atomically rename them.
- [ ] Treat an existing verified content path as idempotent success.
- [ ] Treat an existing path with unexpected size or digest as corruption.
- [ ] Remove only copies created by the current failed operation.
- [ ] Delete copies idempotently and remove empty shard directories when
  practical.
- [ ] Keep source deletion outside the storage backend.
- [ ] Test empty and large files, duplicates, corruption, permissions,
  cancellation, and cleanup.

Acceptance:

- The backend creates, verifies, and deletes physical copies without knowing
  collection schemas or tag rules.

## Phase 17: transactional import workflow

Coordinate catalog, evaluator, repository, and local storages in Core.

- [ ] Define an import request with collection, one source path, and explicit
  tag assignments.
- [ ] Resolve explicit or default collection through the validated snapshot.
- [ ] Reject symlinks, directories, and non-regular sources before mutation.
- [ ] Stream and stage the source once, producing SHA-256 and size.
- [ ] Apply collection-required tags; require explicit values for required
  non-boolean tags.
- [ ] When no storage is explicit or required, use `default_storage_name` only
  if it is available to the collection; otherwise fail before mutation.
- [ ] Assign at least one valid storage using explicit assignments, collection
  requirements, and the documented default rule.
- [ ] Build proposed `FileState` and reject missing demands or active conflicts.
- [ ] Copy or verify content in every selected storage.
- [ ] Commit file, source, typed tags, and storage assignments in one SQLite
  transaction.
- [ ] On copy or SQL failure, remove only new copies, roll back SQL, and preserve
  the source.
- [ ] Run `remove_on_upload` only after every copy and database commit succeeds.
- [ ] Before removing the source, verify it still names the imported bytes.
- [ ] Reject duplicate imports without changing the existing record, copies,
  tags, observed sources, or import source.
- [ ] Return canonical SHA-256 and whether the record/copies were new.

Acceptance:

- Success leaves every requested copy and one consistent database record.
- Failure never commits a partial record or removes the source.

## Phase 18: persisted tag and storage mutation

- [ ] Load persisted files into `FileState`.
- [ ] Implement add/set/remove for every tag type.
- [ ] Canonicalize tag and predefined-value names.
- [ ] Treat boolean false as removal.
- [ ] Enforce collection imports and required tags.
- [ ] Evaluate the complete proposed state before mutation.
- [ ] Return demands, conflicts, suggestions, and originating reasons.
- [ ] Adding `storage:<name>` copies/verifies content before logical commit.
- [ ] Removing `storage:<name>` updates SQLite and deletes exactly that copy
  with recoverable failure reporting.
- [ ] Removing the final storage deletes the file record, tag/source rows, and
  final copy in the documented order.
- [ ] Reject removal of required tags and storages.
- [ ] Make repeated add/remove operations idempotent.
- [ ] Test rollback after copy, delete, evaluator, and database failures.

Acceptance:

- Successful mutations always produce evaluator-valid persisted state.
- Storage tags and physical copies agree after successful calls.

## Phase 19: search

- [ ] Implement the documented query parser independently of Cobra and SQL.
- [ ] Resolve query names through the selected collection catalog.
- [ ] Compile typed equality, presence/absence, range, and membership terms.
- [ ] Combine every term with AND; reject OR, grouping, negated values, fuzzy
  matching, and unsupported operators.
- [ ] Use SQL parameters only; never interpolate user values.
- [ ] Search only tags imported by the selected collection.
- [ ] Return SHA-256, original name, size, and assigned tags needed by clients.
- [ ] Order by `imported_at` descending and full SHA-256 ascending.
- [ ] Implement non-negative limit/offset pagination with defaults of 100 and
  0.
- [ ] Test every type, combined terms, no matches, malformed queries, and
  SQL-injection-shaped input.

Acceptance:

- Search reads persisted SQLite state and shares assignment type rules.

## Phase 20: Core collection session

One process owns at most one open collection session in the MVP.

- [ ] Replace raw database handles returned to frontends with a Core-owned
  session.
- [ ] Open and initialize one validated collection at a time.
- [ ] Bind collection config, catalog subset, graph, evaluator, repository, and
  storage backends into the session.
- [ ] Before publishing the session, validate every persisted tag name, tag
  type, typed value, and storage name against the current catalog and evaluate
  every reconstructed file state.
- [ ] Refuse to open a collection containing incompatible persisted state and
  report the affected SHA-256 values without preventing application startup or
  unrelated collections from opening.
- [ ] Reject opening a second collection before closing the first.
- [ ] Expose import, mutation, lookup, hints, and search through Core.
- [ ] Propagate context cancellation through hashing, copying, SQL, and search.
- [ ] Close the database exactly once on success, failure, and shutdown.
- [ ] Keep frontend types out of Core APIs.
- [ ] Test lifecycle, failed open, repeated close, cancellation, and independent
  valid collections.

Acceptance:

- Frontends cannot bypass validation or coordinate storage and SQL themselves.
- Separate GUI and CLI processes can concurrently use SQLite.

## Phase 21: CLI MVP commands

- [ ] Document arguments, flags, output, and exit behavior before each command.
- [ ] Add `collection <name> import <file>` and the documented default alias.
- [ ] Add explicit `--interactive` import: prompt for missing required tags,
  then offer skippable optional imported tags, and pass a complete request to
  Core.
- [ ] Add collection-specific and default typed tag add/set/remove commands.
- [ ] Add collection-specific and default `tag <sha256> get` commands that
  return every assigned typed tag and storage.
- [ ] Add collection-specific and default search commands.
- [ ] Accept repeated tag assignments during import using Phase 13 syntax.
- [ ] Print canonical SHA-256 after import and mutation.
- [ ] Render evaluator failures and reasons without Cobra usage noise.
- [ ] Add side-effect-free tag/value completion through Core.
- [ ] Use `RunE`, argument validators, `cmd.Context`, and command writers.
- [ ] Test fresh command trees, default/explicit collections, invalid input,
  cancellation, output, and nonzero failures.

Acceptance:

- Every CLI command in `mvp.ai.md` operates on persisted data.
- Explicit collection selection never changes `default_collection`.

## Phase 22: failure, concurrency, and recovery tests

- [ ] Test simultaneous GUI-like and CLI-like database connections.
- [ ] Test busy-timeout behavior and transaction contention.
- [ ] Test source modification during import.
- [ ] Test cancellation during hashing, each copy, and SQL mutation.
- [ ] Test multi-storage failure after each copy.
- [ ] Test SQL failure after physical copies are finalized.
- [ ] Test duplicate concurrent imports of identical content.
- [ ] Test restart handling of orphaned staged/finalized copies under the Phase
  13 crash policy.
- [ ] Test removing one of several copies and the final copy.
- [ ] Test required tags and every relationship against persisted state.
- [ ] Run race tests where supported.

Acceptance:

- No tested failure silently loses the source or commits a row pointing to a
  missing copy.

## Phase 23: MVP documentation and release gate

- [ ] Update `docs/cli.md` with all commands and examples.
- [ ] Update `docs/config-spec.md`, including `remove_on_upload` status.
- [ ] Update `docs/navigation.md` for all new files.
- [ ] Document the dependency flow from config through Core to frontends.
- [ ] Verify clean-home init, check, import, tag, search, storage removal, and
  reopen.
- [ ] Verify all eleven `mvp.ai.md` completion criteria explicitly.
- [ ] Run `go tool task check`.
- [ ] Run relevant race and integration suites.
- [ ] Record platform limitations without weakening data safety.

Acceptance:

- Product behavior, human docs, AI contracts, and tests describe one MVP.
- The release gate cannot pass while persistence or required CLI workflows are
  absent.

## Deferred beyond this TODO

- [ ] Define the database-encryption threat model and key lifecycle: creation,
  acquisition, storage, sharing between GUI and CLI, rotation, and recovery.
- [ ] Enable the `ncruces/go-sqlite3` Adiantum VFS without exposing keys in
  logs, errors, configuration diagnostics, process arguments, or command
  history.
- [ ] Implement and test an interruption-safe plaintext-to-encrypted export and
  replacement workflow; merely reopening a plaintext database with an
  encrypted VFS is not a migration.
- [ ] Keep SQLite temporary data in memory when encryption is enabled and
  document that collection database encryption does not encrypt stored file
  contents.
- Remote/rclone storage.
- Collection tag overrides.
- Relationship targets by group.
- Wildcard imports and exclusions.
- Multiple open collections in one process.
- Persistent GUI window state and live config reload.
