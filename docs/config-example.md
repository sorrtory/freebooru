# FreeBooru configuration example

## Application

```text
$HOME/.config/freebooru/freebooru.yaml
```

```yaml
lang: en
default_collection: main
default_storage_name: default
default_storage_path: $HOME/.local/share/freebooru/storage/default
http_port: 52800
remove_on_upload: false
```

## Storage

```text
$HOME/.config/freebooru/storage.yaml
```

```yaml
- name: default
  type: local
  path: $HOME/.local/share/freebooru/storage/default
- name: archive
  type: local
  path: /mnt/archive/freebooru
```

Storage entries are values of the built-in `storage` tag:

```text
storage:default
storage:archive
```

## Collection

One config per collection. Configs are read recursively from `collections/`.

```text
$HOME/.config/freebooru/collections/main.yaml
```

```yaml
name: main
location: $HOME/.local/share/freebooru/collections/main.sqlite
tags:
    require:
        - storage: default
        - tag: source
    import:
        - storage: archive
        - group: content
        - group: metadata
```

`location` may be omitted. It defaults to
`$HOME/.local/share/freebooru/collections/<name>.sqlite`.

## Tags

Tag configs are read recursively from `tags/`. A file may contain several YAML
documents.

On a new installation, `init` creates the small group-per-file starter catalog
described in [configuration best practices](./config-best_practise.md). The
example below shows a more specialized replacement or extension.

```text
$HOME/.config/freebooru/tags/content.yaml
```

```yaml
name: character
type: multivalue
groups: [content]
suggest:
    - tag: artist
      reason: character images often have a known artist
values:
    - val: cirno
      aliases: [chiruno, ice_fairy, チルノ]
      demand:
          - tag: universe
            has: [touhou]
            reason: cirno is a touhou character
          - tag: baka
            is: true
            reason: cirno is baka
          - tag: rating
            not: [explicit]
          - tag: filesize
            min: 1
            max: 10485760
          - tag: created_on
            after: "2000-01-01"
            before: "2030-01-01"
          - tag: imported_at
            after: "2020-01-01T00:00:00Z"
          - tag: source
            regex: '^https?://'
      conflict:
          - tag: species
            is: human
            reason: cirno is not human
    - val: reimu
---
name: source
type: text
groups: [metadata]
---
name: filesize
type: int
groups: [metadata, technical]
---
name: created_on
type: date
groups: [metadata]
---
name: imported_at
type: datetime
groups: [metadata]
---
name: reviewed
type: bool
groups: [workflow]
---
name: artist
type: text
groups: [content]
---
name: universe
type: multivalue
groups: [content]
values:
    - val: touhou
---
name: baka
type: bool
groups: [content]
---
name: rating
type: value
groups: [content]
values:
    - val: safe
    - val: explicit
---
name: species
type: value
groups: [content]
values:
    - val: human
    - val: fairy
```

Files and folders only organize the config. Tag and group names come from YAML.
Using tag names as filenames and group names as folders is preferred for
readability but has no semantic effect.

`suggest`, `demand`, and `conflict` may be set on a whole tag or under one
predefined value. Relationships are one-way. Their targets must be imported or
required by the collection; the relationship itself does not import a tag.
