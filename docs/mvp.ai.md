# FreeBooru MVP

This document defines the goal and completion criteria for the first usable
FreeBooru release. The detailed configuration contract belongs in
[config-spec.md](./config-spec.md).

## Goal

The MVP must let a user define a collection, import local files, assign typed
tags, search the collection, and control which local storage contains each
file.

FreeBooru must reject ambiguous configuration and must not silently lose file
data.

## User workflow

```text
initialize configuration
    -> check configuration
    -> open a collection
    -> import files
    -> assign and remove tags
    -> search files by tags
```

The CLI must support at least:

```bash
freebooru-cli config init
freebooru-cli config check
freebooru-cli collection main import <path>
freebooru-cli collection main tag <operation>
freebooru-cli collection main search <query>
```

When `default_collection` is `main`, commands that omit `collection main` use
that collection:

```bash
freebooru-cli tag <operation>
freebooru-cli search <query>
```

`default_collection` is a default, not mutable current-collection state. A GUI
window or CLI command may explicitly select another collection without changing
it.

## Configuration initialization

`freebooru-cli config init` must create the initial configuration without
overwriting existing files:

```text
$HOME/.config/freebooru/freebooru.yaml
$HOME/.config/freebooru/storage.yaml
$HOME/.config/freebooru/collections/main.yaml
$HOME/.config/freebooru/tags/
$HOME/.local/share/freebooru/storage/default/
$HOME/.local/share/freebooru/collections/
```

The default application configuration is:

```yaml
lang: en
default_collection: main
default_storage_name: default
default_storage_path: $HOME/.local/share/freebooru/storage/default
http_port: 52800
remove_on_upload: false
```

The default storage is:

```yaml
- name: default
  type: local
  path: $HOME/.local/share/freebooru/storage/default
```

The default collection is:

```yaml
name: main
location: $HOME/.local/share/freebooru/collections/main.sqlite
tags:
    require:
        - storage: default
```

## Collections

- A collection has one SQLite database.
- A collection is selected explicitly or through `default_collection`.
- One process opens at most one collection in the MVP.
- A collection must allow at least one storage.
- `location` is optional and defaults to
  `$HOME/.local/share/freebooru/collections/<name>.sqlite`.
- A collection cannot open while its referenced storage or tag configuration
  is invalid.
- An invalid collection must not prevent unrelated valid collections from
  opening.

SQLite coordinates ordinary simultaneous access by the GUI and CLI. General
collection lockfiles and multiple open collections in one process are outside
the MVP.

## Storage and files

The MVP supports local storage only.

Storage is exposed as the built-in multivalue tag `storage`:

```text
storage:default
storage:archive
```

- Assigning a storage value copies the file into that storage.
- Removing a storage value deletes the copy from that storage.
- Removing the last storage value deletes the indexed file and its physical
  data.
- Assigning multiple storage values creates one copy in every selected storage.
- A collection may require a storage value like any other required tag.
- The collection database stores the file index, assigned tags, and storage
  locations. File contents remain in storage.
- `remove_on_upload` removes the source only after every requested copy and the
  database transaction have succeeded.
- If a multi-storage import fails, FreeBooru must remove copies created by that
  operation, leave the source intact, and not commit the database record.

## Tags

The MVP supports these tag types:

| Type         | Value                                                     |
| ------------ | --------------------------------------------------------- |
| `bool`       | Presence is true; absence is false                        |
| `text`       | Free text                                                 |
| `int`        | Integer from `0` through the SQLite signed 64-bit maximum |
| `date`       | ISO date in `YYYY-MM-DD` form                             |
| `datetime`   | RFC 3339 timestamp                                        |
| `value`      | One predefined value                                      |
| `multivalue` | Zero or more predefined values                            |

For boolean constraints, `is: false` means that the tag must be absent. False
is not stored as an assigned tag value.

A tag may belong to multiple implicit groups:

```yaml
groups:
    - metadata
    - technical
```

Group and tag names may be equal because references identify whether they mean
a tag or a group.

Within a collection:

- `import` makes a tag available.
- `require` implies `import` and requires the tag on every indexed file.
- Any tag mentioned by the collection becomes available to that collection.
- Tag overrides are outside the MVP.

### Tag relationships

Tags and predefined values may declare directed `suggest`, `demand`, and
`conflict` relationships. Tag-level rules apply while the source tag exists;
value-level rules apply while that value is assigned.

- `suggest` recommends a target without invalidating the file.
- `demand` requires a matching target.
- `conflict` forbids a matching target.
- No predicate means that the target tag must be present.
- `has` and `not` test predefined values; `is` tests an exact typed value.
- `min` and `max` are inclusive integer bounds.
- `before` and `after` are strict date or datetime bounds.
- `regex` uses Go regular-expression matching for text.

`min`+`max`, `before`+`after`, and `has`+`not` may be combined. `is` is
exclusive with every other predicate. For booleans, `is: false` means absence.

Relationship targets must be imported or required by the collection; graph
edges do not auto-import tags. `suggest` and `demand` cycles are valid.
Self-demand, self-conflict, and demanding and conflicting with the same target
condition from the same source condition are invalid.

## Names and case

Collection, storage, tag, group, type, and predefined-value names are
case-insensitive. Filesystem paths and free text remain case-sensitive.

Names must follow [config-best_practise.md](./config-best_practise.md):

- non-empty;
- no leading or trailing whitespace;
- Unicode letters, digits, and `_` only;
- no spaces or punctuation;
- `:` separates a tag name from its value;
- `-` is reserved;
- unique within their category after case normalization.

Duplicate definitions invalidate every conflicting definition. File discovery
order must not decide which definition wins.

## Paths and YAML

- `$HOME` and `~` are expanded.
- Relative paths are invalid.
- `freebooru.yaml` and `storage.yaml` contain one YAML document each.
- Every collection file contains one YAML document.
- Tag files may contain multiple YAML documents.
- Tags and collections are discovered recursively from `.yaml` and `.yml`
  files.
- Other files are ignored.
- Empty YAML files and unknown fields are invalid.
- Symlinked directories are not followed.

## Error behavior

- Invalid or missing `freebooru.yaml` prevents normal application operation.
- Invalid storage, tag, or collection configuration does not prevent the
  application from starting and reporting the problem.
- Invalid definitions cannot be used.
- A collection with an invalid dependency cannot open.
- Unrelated valid collections remain usable.
- `freebooru-cli config check` reports every discoverable configuration problem
  and exits with a nonzero status when any problem exists.

## Out of scope

The MVP does not include:

- rclone or other remote storage;
- tag overrides in collection configuration;
- wildcard group imports or exclusions;
- multiple open collections in one process;
- persistent GUI window state;
- application-level state SQLite;
- general collection lockfiles;
- live configuration reload.

These features must not weaken the explicit collection, storage, and tag
contracts defined for the MVP.

## Completion criteria

The MVP is complete when:

1. `config init` creates a usable default configuration without overwriting
   existing configuration.
2. `config check` detects malformed YAML, invalid fields, duplicates, and broken
   references.
3. The default collection can be opened and its SQLite schema initialized.
4. A local file can be imported transactionally into one or more local
   storages.
5. Typed tags can be assigned, removed, persisted, and queried.
6. Required tags are enforced for every indexed file.
7. Storage-tag changes copy and delete physical file copies consistently.
8. The GUI and CLI can access the same collection through SQLite without using
   global current-collection state.
9. Automated tests cover configuration validation, import rollback, tag
   persistence, search, and storage-copy deletion.
10. Relationship validation and enforcement cover every MVP predicate and the
    graph rules defined above.
11. `go tool task check` succeeds.
