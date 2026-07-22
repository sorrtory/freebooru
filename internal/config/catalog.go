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
	groups      map[string]Group
	references  map[string]ResolvedReferences
}

// Group is an implicit navigation group formed by tag memberships.
type Group struct {
	Name string
	tags []string
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
