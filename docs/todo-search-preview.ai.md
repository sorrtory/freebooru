# Search completion and text preview contract

- [x] Header and Files search share collection-aware tag and persisted-value hints.
- [x] Tag matching is case-insensitive substring matching; value matching uses the active `tag:value` prefix.
- [x] Enter in header search always opens Files with the entered query.
- [x] Presence queries such as `artist` list files having that assignment.
- [x] Collection and search popovers close on outside interaction.
- [x] Popular tags toggle their exact query term instead of accumulating duplicates.
- [x] The whole imported tag card opens Files with that tag-presence query.
- [x] Text-like file content has a bounded, read-only Monaco preview with an Open file fallback.

SQLite FTS is not required for this phase. Tag catalogs are configuration-sized
and use normalized substring matching. Persisted values are collected through
the collection repository with bounded, distinct results; an FTS
index remains an optimization option after profiling larger collections.
