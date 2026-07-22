# Configuration best practices

## Booru style guide

- [How to tag](https://thewanderinginn.fandom.com/wiki/Booru_Tagging_Guideline#How_to_tag)
- [What to tag](https://thewanderinginn.fandom.com/wiki/Booru_Tagging_Guideline#What_to_tag)
- [Danbooru tag categories](https://safebooru.donmai.us/wiki_pages/help%3Atags)
- [Danbooru post relationships](https://safebooru.donmai.us/wiki_pages/help%3Apost_relationships)
- [Hydrus tag parents](https://hydrusnetwork.github.io/hydrus/advanced_parents.html)

## Convention

- case-insensitive: names ; case-sensitive: paths, free text
- names are globally unique within their category (tag, group, collection, storage)
- non-empty
- no leading/trailing whitespace
- naming characters: unicode letters (a-z, а-я), digits, `_`. nothing else. no spaces, no special characters, no punctuation. 
- `:` - is used to separate tag name and value, e.g. `character:cirno`
- `-` - is reserved

Tag what is observable or known, not an opinion inferred from the file. Prefer
one canonical spelling for a concept. Aliases and implication rules should be
used to remove repetition; adding several synonymous tags to every file is not
a substitute for either feature.

## Default tag catalog

`freebooru-cli init` creates one YAML file per primary group. This is the
preferred default layout because it is easy to browse and edit:

| File | Group | Initial tags |
| --- | --- | --- |
| `tags/creator.yaml` | `creator` | `artist` |
| `tags/universe.yaml` | `universe` | `universe` |
| `tags/character.yaml` | `character` | `character` |
| `tags/general.yaml` | `general` | `rating`, `description` |
| `tags/metadata.yaml` | `metadata` | `source` |

The filename is organization for humans; `groups` in each tag document is the
semantic group membership. One file per group is a good default, but a tag may
belong to multiple groups and larger catalogs may use nested directories.

The starter `universe` and `character` values are deliberately small. Replace
or extend them with the collection's real controlled vocabulary. A character
value should normally demand its universe once that relationship is known.

### Core metadata is not ordinary tag configuration

Every indexed file must have core-owned metadata. Users do not assign these
values and they should not be duplicated as YAML tags:

| Property | Ownership | Current status |
| --- | --- | --- |
| SHA-256 | calculated from the imported bytes | stored |
| file size in bytes | calculated while importing | stored |
| upload/import time | set by the collection database | stored |
| record creation and update times | set by the collection database | stored |
| original path and filename | captured from the import source | stored |
| file/MIME type | inspect the bytes, never trust only the suffix | planned |
| last interaction time | update on a clearly defined user interaction | planned |

This metadata is required by the data model even though it is not listed under
the collection's `tags.require`. Search should eventually expose it through
system predicates such as `sha256:`, `filesize:`, and `imported_at:`. The
`metadata` YAML group is reserved for descriptive, user-managed provenance such
as `source`; it does not shadow system properties.

## Versions and related files

Do not represent versions with a plain counter tag. `version:3` has no stable
meaning when a version is deleted, discovered late, or when two edits branch
from the same original. A linked list also makes sibling discovery and choosing
the best representative unnecessarily expensive.

Use a file relationship instead: each variant stores one optional parent file
identifier, while the database can query all children. This forms a shallow
tree, not an ordered linked list. Keep a separate relationship kind such as
`variant`, `alternate`, or `derived`, and optionally store a user-controlled
display order. The parent should be the preferred or canonical file, not
necessarily the oldest upload. Ordered works such as comic pages belong in a
collection/pool with an explicit position rather than a version family.

File relationships are not part of the current configuration or database
specification. They should be designed as first-class database records rather
than encoded as tag text.
