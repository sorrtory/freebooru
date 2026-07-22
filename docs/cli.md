# freebooru-cli

A tool to manage FreeBooru collections and tags from the command line.
It calls the FreeBooru core.

## Commands

All commands use the configured `default_collection` unless they begin with
`collection <name>`. Explicit collection selection never changes that default.

| Command                              | Arguments and flags                                       | Standard output                                  |
| ------------------------------------ | --------------------------------------------------------- | ------------------------------------------------ |
| `init`                               | no arguments                                              | `FreeBooru initialized`                          |
| `config check`                       | no arguments                                              | `Configuration is valid`                         |
| `collection list`                    | no arguments                                              | one collection name per line                     |
| `collection edit <name>`             | no flags                                                  | no output                                        |
| `storage list`                       | no arguments                                              | one imported storage name per line               |
| `storage edit`                       | no arguments                                              | no output                                        |
| `tag list`                           | no arguments                                              | one imported tag name per line                   |
| `tag edit <name>`                    | no flags                                                  | no output                                        |
| `files list`                         | optional `--storage <name>`                               | `sha256<TAB>storage` per assignment               |
| `import <file>`                      | repeatable `--tag name[:value]`; optional `--interactive` | imported SHA-256                                 |
| `collection <name> import <file>`    | same as `import`                                          | imported SHA-256                                 |
| `tag <sha256> add <name[:value]>`    | no flags                                                  | canonical SHA-256                                |
| `tag <sha256> set <name[:value]>`    | no flags                                                  | canonical SHA-256                                |
| `tag <sha256> remove <name[:value]>` | a value is required only for `storage`                    | canonical SHA-256                                |
| `tag <sha256> get`                   | no flags                                                  | one canonical `name[:value]` assignment per line |
| `collection <name> tag ...`          | same operations as `tag`                                  | same as `tag`                                    |
| `search <term>...`                   | `--limit` defaults to `100`; `--offset` defaults to `0`   | one matching SHA-256 per line                    |
| `collection <name> search <term>...` | same as `search`                                          | same as `search`                                 |
| `collection <name> storage list`     | no arguments                                              | same as `storage list`                           |
| `collection <name> tag list/edit`    | same as `tag list/edit`                                   | same as `tag list/edit`                          |
| `collection <name> files list`       | optional `--storage <name>`                               | same as `files list`                             |

Repeated `--tag` flags preserve their order. A boolean assignment is written
as `name`; every other type uses `name:value`. The first colon separates the
tag name, so later colons remain part of the value. A multivalue tag may be
repeated; assigning the same scalar tag more than once is an error.

`tag get` prints multivalue assignments as one line per value and includes
storage assignments as `storage:<name>`. Output ordering is deterministic by
normalized tag name and then value.

List output is deterministic. `storage list`, `tag list`, and `files list`
use the selected collection. `collection list` is global. `storage edit` opens
the global `storage.yaml`; `collection edit` and `tag edit` open the YAML file
that defines the selected entry. The editor is taken from `$EDITOR`, is run
directly without a shell, and must exit before the command continues. Edited
configuration is validated after the editor exits; invalid edits make the
command fail.

`files list` prints one row for every file/storage assignment. A file stored in
two storages therefore appears twice. `--storage` accepts only a storage
imported by the selected collection and restricts the rows to that storage.

Interactive import prompts on standard error and reads standard input. It asks
for missing required assignments first, then offers optional imported tags;
the optional section may be skipped. Boolean answers accept `true`/`false` or
`yes`/`no`; multivalue answers are comma-separated. It produces the same Core
import request as non-interactive import.

Successful data commands write only their documented result to standard
output. Diagnostics and verbose logs go to standard error. Invalid arguments,
configuration, evaluator failures, cancellation, and persistence failures
return a nonzero status without printing Cobra usage. `--help` is the only
normal path that prints usage.

## Examples

Initialize and validate a clean installation:

```bash
freebooru-cli init
freebooru-cli config check
freebooru-cli collection list
freebooru-cli storage list
freebooru-cli tag list
```

Import into the configured default collection and inspect the result:

```bash
sha=$(freebooru-cli import ./image.png --tag reviewed --tag rating:safe)
freebooru-cli tag "$sha" get
freebooru-cli search reviewed rating:safe
```

Run the same workflows against an explicit collection without changing
`default_collection`:

```bash
sha=$(freebooru-cli collection archive import ./image.png --interactive)
freebooru-cli collection archive tag "$sha" set rating:questionable
freebooru-cli collection archive search rating:questionable --limit 25
freebooru-cli collection archive files list --storage archive
```

Search terms are combined with AND. Presence, absence, scalar equality,
multivalue membership, and ordered comparisons are supported:

```bash
freebooru-cli search reviewed
freebooru-cli search '!blocked' rating:safe labels:portrait
freebooru-cli search 'score>=10'
freebooru-cli search 'published>=2026-01-01'
freebooru-cli search 'captured<2026-07-23T12:00:00Z'
```

Storage uses the tag command surface. Adding a value copies content; removing
one deletes that copy. Removing the last value also deletes the indexed file:

```bash
freebooru-cli tag "$sha" add storage:archive
freebooru-cli tag "$sha" remove storage:default
```

There is no `config init`: initialization is an application workflow, not a
configuration-only operation. Successful initialization prints
`FreeBooru initialized`. A failure returns a nonzero status and does not print
the success message.
