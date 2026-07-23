# FreeBooru MVP config contract

Implementation-facing summary. Human documentation lives in
[config-spec.md](./config-spec.md) and [config-example.md](./config-example.md).

## Files

- Required singleton documents: `freebooru.yaml`, `storage.yaml`.
- Recursively load `.yaml` and `.yml` under `collections/` and `tags/`.
- One document per collection file; multiple documents allowed per tag file.
- Ignore other files and do not follow symlinked directories.
- Reject empty documents, unknown fields, and duplicate normalized names.

## Normalization and paths

- Case-insensitive: collection, storage, tag, group, type, and predefined-value
  names.
- Case-sensitive: filesystem paths and free text.
- Names: Unicode letters, digits, `_`; non-empty; no surrounding whitespace.
- Expand `$HOME` and leading `~`; reject relative paths.

## Application

```yaml
lang: en
default_collection: main
default_storage_name: default
default_storage_path: $HOME/.local/share/freebooru/storage/default
http_address: 0.0.0.0
http_port: 52800
remove_on_upload: false
```

- Missing or invalid application config blocks normal operation.
- Explicit collection/storage selection overrides defaults.
- `remove_on_upload` runs only after all copies and the DB transaction succeed.

## Storage

MVP provider: `{name, type: local, path, comment?}`. `name` becomes a value of the built-in
multivalue tag `storage`. Assigning/removing values copies/deletes physical
copies. Removing the last value deletes the indexed file. Multi-storage failure
rolls back new copies and the DB record and preserves the upload source.

## Collection

Fields: `name` required, `location` and `comment` optional, `tags` required. Default location:
`$HOME/.local/share/freebooru/collections/<name>.sqlite`.

`tags.require` and `tags.import` contain objects with exactly one of `tag`,
`group`, or `storage`. At least one storage reference is required. `require`
implies and overrides `import`. Missing/invalid references make the collection
unopenable without disabling unrelated collections.

## Tags

Fields: `name`, `type`, optional `comment`, optional `groups`, and conditional
`values`. Each predefined value may also have `comment`. `storage` and the
system metadata names are reserved. Comments must be non-blank when present and
do not affect validation. Groups are implicit and many-to-many; group/tag names
may collide.

Reserved read-only system tags are `sha256`, `filesize`, `filetype`,
`imported_at`, `updated_at`, and `last_interaction_at`. They resolve directly to
authoritative `file` columns for search and are never accepted as assignments.

- `bool`: presence=true, absence=false.
- `text`: case-sensitive string.
- `int`: `0..9223372036854775807`.
- `date`: `YYYY-MM-DD`.
- `datetime`: RFC 3339.
- `value`: exactly one configured canonical `val`.
- `multivalue`: zero or more unique configured canonical `val` values.

Each predefined value may declare `aliases`. Canonical values and aliases are
case-insensitively unique within their tag and use the normal name syntax.
Assignment, search, and relationship predicate inputs accept either spelling;
Core and the graph resolve aliases to `val` before evaluation or persistence.
Aliases are never stored in SQLite and never create separate graph nodes.

`suggest`, `demand`, and `conflict` are MVP. They may occur at tag level or
inside a predefined value. Tag-level rules activate on tag presence; value-level
rules activate only for that value. Edges are directed and never imply a reverse
edge.

- `suggest`: recommendation only.
- `demand`: matching target is required.
- `conflict`: matching target is forbidden with the source.
- Relationship `tag` is required.
- No predicate means target presence.
- `has: [...]`: multivalue target contains at least one listed value.
- `is: x`: target equals `x`; boolean false means tag absence.
- `not: [...]`: value/multivalue target contains none of the listed values.
- `min: n`, `max: n`: inclusive bounds for an `int` target.
- `before: x`, `after: x`: strict bounds for a `date` or `datetime` target.
- `regex: x`: Go regexp matched against a `text` target; substring semantics
  apply unless anchored.
- `reason`: optional non-empty human explanation.

`min`+`max`, `before`+`after`, and `has`+`not` may be combined. `is` is
exclusive with every other predicate. A target must be available through the
collection's `import` or `require`; relationships never auto-import tags.

`suggest` and `demand` cycles are valid. Self-demand and self-conflict are
invalid. For the same source condition, demanding and conflicting with the same
target condition is invalid. Group relationship targets, collection tag
overrides, rclone, wildcard imports, and exclusions are outside MVP.

## Failure policy

Start the application when storage/tag/collection configs are broken, report
all discoverable errors, and prohibit use of invalid definitions. `config check`
returns nonzero when any error exists.
