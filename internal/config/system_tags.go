package config

// SystemTags returns read-only tag-shaped views of authoritative file metadata.
func SystemTags() []TagConfig {
	return []TagConfig{
		{Name: "filetype", Type: TagTypeText, Comment: "Detected MIME type"},
		{Name: "filesize", Type: TagTypeInt, Comment: "File size in bytes"},
		{Name: "imported_at", Type: TagTypeDatetime, Comment: "Collection import time"},
		{
			Name: "last_interaction_at", Type: TagTypeDatetime,
			Comment: "Last explicit metadata or relationship mutation",
		},
		{Name: "sha256", Type: TagTypeText, Comment: "Lowercase SHA-256 content identity"},
		{Name: "updated_at", Type: TagTypeDatetime, Comment: "Last record update time"},
	}
}

// SystemTag resolves one reserved system metadata name case-insensitively.
func SystemTag(name string) (TagConfig, bool) {
	key := normalizeName(name)
	for _, tag := range SystemTags() {
		if normalizeName(tag.Name) == key {
			return tag, true
		}
	}
	return TagConfig{}, false
}
