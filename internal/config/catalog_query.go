package config

import (
	"sort"
	"strings"
)

// Storage returns a storage definition using case-insensitive lookup.
func (c *Catalog) Storage(name string) (StorageProvider, Source, bool) {
	entry, ok := c.storages[normalizeName(name)]
	return entry.value, entry.source, ok
}

// Tag returns a defensive copy using case-insensitive lookup.
func (c *Catalog) Tag(name string) (TagConfig, Source, bool) {
	entry, ok := c.tags[normalizeName(name)]
	if !ok {
		return TagConfig{}, Source{}, false
	}
	return cloneTag(entry.value), entry.source, true
}

// Group returns an implicit group using case-insensitive lookup.
func (c *Catalog) Group(name string) (Group, bool) {
	group, ok := c.groups[normalizeName(name)]
	return group, ok
}

// TagsInGroup returns defensive tag copies in normalized-name order.
func (c *Catalog) TagsInGroup(name string) []TagConfig {
	group, ok := c.groups[normalizeName(name)]
	if !ok {
		return nil
	}
	tags := make([]TagConfig, 0, len(group.tags))
	for _, key := range group.tags {
		tags = append(tags, cloneTag(c.tags[key].value))
	}
	return tags
}

// DeclaredValues returns defensive copies of a tag's predefined values.
func (c *Catalog) DeclaredValues(tagName string) []PredefinedValue {
	entry, ok := c.tags[normalizeName(tagName)]
	if !ok {
		return nil
	}
	return cloneValues(entry.value.Values)
}

// SearchTags returns defensive copies whose normalized names start with prefix.
func (c *Catalog) SearchTags(prefix string) []TagConfig {
	prefix = normalizeName(prefix)
	keys := make([]string, 0)
	for key := range c.tags {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	tags := make([]TagConfig, 0, len(keys))
	for _, key := range keys {
		tags = append(tags, cloneTag(c.tags[key].value))
	}
	return tags
}

// Collection returns a defensive copy using case-insensitive lookup.
func (c *Catalog) Collection(name string) (CollectionConfig, Source, bool) {
	entry, ok := c.collections[normalizeName(name)]
	if !ok {
		return CollectionConfig{}, Source{}, false
	}
	value := entry.value
	value.Tags.Require = append([]TagReference(nil), value.Tags.Require...)
	value.Tags.Import = append([]TagReference(nil), value.Tags.Import...)
	return value, entry.source, true
}

// CollectionReferences returns defensive copies of effective references.
func (c *Catalog) CollectionReferences(name string) (ResolvedReferences, bool) {
	references, ok := c.references[normalizeName(name)]
	if !ok {
		return ResolvedReferences{}, false
	}
	references.Required = append([]TagReference(nil), references.Required...)
	references.Imported = append([]TagReference(nil), references.Imported...)
	return references, true
}
