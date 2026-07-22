# FreeBooru full config specification

See [configuration examples](./config-example.md) and
[configuration best practices](./config-best_practise.md).

## `freebooru.yaml`

This is the application config. FreeBooru cannot operate normally when it is
missing or invalid.

| Field                  | Type   | Required | Default                                        | Rules                                   | Description                                      | Status          |
| ---------------------- | ------ | -------: | ---------------------------------------------- | --------------------------------------- | ------------------------------------------------ | --------------- |
| `lang`                 | enum   |       no | `en`                                           | `en`, `ru`                              | Application language                             | Not implemented |
| `default_collection`   | string |       no | `main`                                         | Existing collection name                | Collection used when none is selected explicitly | Not implemented |
| `default_storage_name` | string |       no | `default`                                      | Valid storage name                      | Storage used by default                          | Not implemented |
| `default_storage_path` | path   |       no | `$HOME/.local/share/freebooru/storage/default` | Absolute after expanding `$HOME` or `~` | Path created for the default storage             | Not implemented |
| `http_port`            | int    |       no | `52800`                                        | `1`–`65535`                             | HTTP server port                                 | Not implemented |
| `remove_on_upload`     | bool   |       no | `false`                                        | —                                       | Remove the source after a successful upload      | Not implemented |

`default_collection` is a fallback, not global current state. An explicitly
selected collection always wins.

## `storage.yaml`

This defines storage backends. Storage is also the built-in multivalue tag
`storage`, for example `storage:default`.

| Field  | Type   | Required | Default | Rules                                   | Description             | Status          |
| ------ | ------ | -------: | ------- | --------------------------------------- | ----------------------- | --------------- |
| `name` | string |      yes | —       | Non-empty and globally unique           | Storage tag value       | Not implemented |
| `type` | enum   |      yes | —       | `local` for MVP                         | Storage implementation  | Not implemented |
| `path` | path   |      yes | —       | Absolute after expanding `$HOME` or `~` | Local storage directory | Not implemented |

Assigning a storage value copies a file there. Removing it deletes that copy.
Removing the last storage value deletes the indexed file. Multi-storage uploads
must roll back if any copy fails.

TODO: add rclone support.

## Collections

Collection configs are read recursively from `collections/`. Each file contains
one collection and each collection has one SQLite database.

| Field      | Type   | Required | Default                                                  | Rules                                                                   | Description                      | Status          |
| ---------- | ------ | -------: | -------------------------------------------------------- | ----------------------------------------------------------------------- | -------------------------------- | --------------- |
| `name`     | string |      yes | —                                                        | Non-empty and globally unique                                           | Collection name                  | Not implemented |
| `location` | path   |       no | `$HOME/.local/share/freebooru/collections/<name>.sqlite` | Absolute after expansion                                                | SQLite database path             | Not implemented |
| `tags`     | object |      yes | —                                                        | Contains `require` and/or `import`; must reference at least one storage | Tags available in the collection | Not implemented |

### Collection tags

| Field     | Type | Required | Default | Rules                 | Description                 | Status          |
| --------- | ---- | -------: | ------- | --------------------- | --------------------------- | --------------- |
| `require` | list |       no | empty   | References must exist | Tags required on every file | Not implemented |
| `import`  | list |       no | empty   | References must exist | Tags available to users     | Not implemented |

Every entry references exactly one `tag`, `group`, or `storage`. `require`
automatically imports the reference and has priority over `import`. Overrides
are not part of the MVP.

## Tags

Tag configs are read recursively from `tags/`. One file may contain multiple
YAML documents separated by `---`.

| Field      | Type   |                  Required | Default | Rules                                              | Description                                  | Status          |
| ---------- | ------ | ------------------------: | ------- | -------------------------------------------------- | -------------------------------------------- | --------------- |
| `name`     | string |                       yes | —       | Non-empty and globally unique; cannot be `storage` | Tag name                                     | Not implemented |
| `type`     | enum   |                       yes | —       | See tag types below                                | Tag value type                               | Not implemented |
| `groups`   | list   |                        no | empty   | No duplicates inside one tag                       | Implicit groups containing this tag          | Not implemented |
| `values`   | list   | for `value`, `multivalue` | —       | Entries contain unique `val` fields                | Allowed predefined values                    | Not implemented |
| `suggest`  | list   |                        no | empty   | See tag relationships                              | Tags recommended with this tag                | Not implemented |
| `demand`   | list   |                        no | empty   | See tag relationships                              | Tags required with this tag                   | Not implemented |
| `conflict` | list   |                        no | empty   | See tag relationships                              | Tags forbidden together with this tag         | Not implemented |

Groups are created implicitly. A tag may belong to several groups. Tag and
group names may be equal because references identify their category.

| Type         | Value                                          |
| ------------ | ---------------------------------------------- |
| `bool`       | Present is true; absent is false               |
| `text`       | Case-sensitive text                            |
| `int`        | Integer from `0` through `9223372036854775807` |
| `date`       | `YYYY-MM-DD`                                   |
| `datetime`   | RFC 3339 timestamp                             |
| `value`      | One predefined value                           |
| `multivalue` | Zero or more predefined values                 |

Each `values` entry contains `val` and may contain its own `suggest`, `demand`,
and `conflict` lists. A tag-level relationship applies whenever the tag exists.
A value-level relationship applies only when that value is assigned.

### Tag relationships

Relationships are one-way. A relationship from tag A to tag B does not create
the reverse relationship.

- `suggest` recommends a matching target but does not make a file invalid.
- `demand` requires a matching target.
- `conflict` makes the source and matching target invalid together.

| Field    | Type   | Required | Default | Rules                                  | Description                                      | Status          |
| -------- | ------ | -------: | ------- | -------------------------------------- | ------------------------------------------------ | --------------- |
| `tag`    | string |      yes | —       | Existing tag name                      | Relationship target                              | Not implemented |
| `has`    | list   |       no | —       | Target type is `multivalue`            | Target contains at least one listed value        | Not implemented |
| `is`     | scalar |       no | —       | Value matches target type              | Target equals this value                         | Not implemented |
| `not`    | list   |       no | —       | Target type is `value` or `multivalue` | Target contains none of the listed values        | Not implemented |
| `min`    | int    |       no | —       | Target type is `int`                    | Target is greater than or equal to this value     | Not implemented |
| `max`    | int    |       no | —       | Target type is `int`                    | Target is less than or equal to this value        | Not implemented |
| `before` | string |       no | —       | Target type is `date` or `datetime`     | Target is earlier than this date or timestamp     | Not implemented |
| `after`  | string |       no | —       | Target type is `date` or `datetime`     | Target is later than this date or timestamp       | Not implemented |
| `regex`  | string |       no | —       | Target type is `text`; valid Go regexp  | Target text matches the regular expression        | Not implemented |
| `reason` | string |       no | —       | Non-empty when present                 | Human explanation for the relationship           | Not implemented |

With no predicate, the relationship matches when the target tag exists. For a
boolean target, `is: true` means present and `is: false` means absent.

Compatible range predicates may be combined: `min` with `max`, `before` with
`after`, and `has` with `not`. `is` cannot be combined with another predicate.
Bounds set by `min` and `max` are inclusive; `before` and `after` are strict.
`regex` uses Go regular-expression matching, so it may match part of the text
unless the expression is anchored.

Every relationship target must already be made available by the collection
through `import` or `require`. A relationship does not import its target.

`suggest` and `demand` cycles are allowed. A rule cannot demand or conflict
with itself. The same source condition cannot both demand and conflict with the
same target condition.

TODO: add relationship targets by group and collection tag overrides.

## Common behavior

- Identifiers and predefined values are case-insensitive. Paths and free text
  are case-sensitive.
- Names follow [config-best_practise.md](./config-best_practise.md).
- Relative paths are not supported.
- Unknown YAML fields, empty YAML files, and duplicate names are invalid.
- `.yaml` and `.yml` files are supported. Other files are ignored.
- Invalid domain configs are reported without stopping the application, but an
  invalid definition cannot be used and a collection with an invalid dependency
  cannot be opened.
