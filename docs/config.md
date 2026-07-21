## Storage provider

This is per device (freebooru instance). couse we need to register folders and rclone etc

```
$HOME/.config/freebooru/storage.yaml
```

```yaml
- name: folder-in-downloads
  type: local
  path: $HOME/Downloads/freebooru/
- name: rclone-yandex-disk
  type: rclone
  tokenfile: $HOME/.config/rclone/rclone.conf
```

## Collection config example

Per collection

```
$HOME/.config/freebooru/collections/<collection-name>.yaml
```

```yaml
name: main # equal to filename
location: $HOME/.local/share/freebooru/main.sqlite
storage: # from storage.yaml
  - folder-in-downloads
  - rclone-yandex-disk
tags:
  require:
    - tag: tag-name
      ...overrides
  import:
    - tag: tag-name
      ...overrides
```

## Tag config example

```
$HOME/.config/freebooru/tags/<tag-name>.yaml
```

```yaml
name: character # <tag-name>
type: multivalue
values:
    - val: cirno
      suggest:
          - tag: color
            has: [blue]
            reason: cirno is usually depicted as a blue fairy
      demand:
          - tag: universe
            has: [touhou] # universe has to be touhou at least
            reason: cirno is a character from touhou
          - tag: baka
            reason: cirno is baka
      conflict:
          # - tag: species
          #   has: [human] # cirno is not human, but pic can contain human too!
          - tag: power_level
            less: 100
            reason: cirno is the strongest 100/100
```

tags could also be grouped

```
$HOME/.config/freebooru/tags/metadata.yaml
```

```yaml
name: filesize
type: int # bytes
group: metadata
---
name: resolution
type: int # pixels
group: metadata
...
```

then import all tags inside a group in collection config

```yaml
tags:
    import:
        - group: metadata
```

sometimes spliting into subfolders by group is more convenient

```
$HOME/.config/freebooru/tags/touhou/characters-a-f.yaml
```

Basically, the config is a flat bunch of yaml objects traversed recursively.
Meaning the config is flexible and you can organize it as you like e.g.

- filename = tag name / group name
- folder = group name


---

## Full config semantics

```yaml

```



