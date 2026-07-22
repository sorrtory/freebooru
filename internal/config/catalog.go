package config

import "strings"

// Source identifies the YAML document that defines a catalog entry.
type Source struct {
	File     string
	Document int
}

// catalogEntry keeps a validated value together with the document that
// declared it. The source is used to produce useful cross-file diagnostics.
type catalogEntry[T any] struct {
	value  T
	source Source
}

// Catalog is the usable, cross-checked view of all domain configuration.
//
// YAML structs describe individual files. Catalog connects those files: it
// indexes definitions by case-insensitive name, rejects ambiguous duplicates,
// and resolves collection references. Core and future autocomplete or graph
// code can query this snapshot without reading YAML again.
//
// Broken definitions are reported as diagnostics and omitted from their index,
// while unrelated valid definitions remain available. Its maps remain private
// and lookup methods return copies so callers cannot mutate the snapshot.
type Catalog struct {
	storages    map[string]catalogEntry[StorageProvider]
	collections map[string]catalogEntry[CollectionConfig]
	tags        map[string]catalogEntry[TagConfig]
	references  map[string]ResolvedReferences
}

// ResolvedReferences is the effective view of one collection's references.
// Required contains references mandatory on every file. Imported contains all
// references available to the collection, including Required because requiring
// a tag or storage necessarily makes it available.
type ResolvedReferences struct {
	Required []TagReference
	Imported []TagReference
}

func normalizeName(name string) string {
	return strings.ToLower(name)
}

// Storage returns a storage definition using case-insensitive lookup.
func (c *Catalog) Storage(name string) (StorageProvider, Source, bool) {
	entry, ok := c.storages[normalizeName(name)]
	return entry.value, entry.source, ok
}

// Tag returns a tag definition using case-insensitive lookup.
func (c *Catalog) Tag(name string) (TagConfig, Source, bool) {
	entry, ok := c.tags[normalizeName(name)]
	return entry.value, entry.source, ok
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
