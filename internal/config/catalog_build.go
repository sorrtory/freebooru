package config

import "fmt"

// LoadCatalog turns the domain YAML files into one queryable snapshot.
//
// The build has three stages:
//  1. Decode and locally validate each storage, tag, and collection.
//  2. Normalize names and index definitions, excluding every duplicate.
//  3. Resolve cross-file references and application defaults.
//
// It returns a catalog even when diagnostics contain errors. This lets the
// application start and use independent valid definitions while showing which
// configurations are unavailable and why.
func LoadCatalog(paths Paths, app AppConfig) (*Catalog, Diagnostics) {
	// Loading preserves valid documents even when another document is broken.
	storages, diagnostics := loadStorages(paths.Storage)
	tags, tagDiagnostics := loadYAMLDir(paths.Tags, "tag", true, VerifyTagConfig)
	diagnostics = append(diagnostics, tagDiagnostics...)
	collections, collectionDiagnostics := loadYAMLDir(
		paths.Collections,
		"collection",
		false,
		VerifyCollectionConfig,
	)
	diagnostics = append(diagnostics, collectionDiagnostics...)

	// Indexing establishes the single normalized identity used by all lookups.
	catalog := &Catalog{}
	catalog.storages, diagnostics = indexDefinitions(
		storages,
		"storage",
		func(provider StorageProvider) string { return provider.Name },
		diagnostics,
	)
	catalog.tags, diagnostics = indexDefinitions(
		tags,
		"tag",
		func(tag TagConfig) string { return tag.Name },
		diagnostics,
	)
	catalog.addStorageTag(paths.Storage)
	catalog.groups = buildGroups(catalog.tags)
	catalog.collections, diagnostics = indexDefinitions(
		collections,
		"collection",
		func(collection CollectionConfig) string { return collection.Name },
		diagnostics,
	)
	// Cross-file checks happen only after all independent definitions are known.
	diagnostics = append(diagnostics, catalog.resolveCollections()...)
	diagnostics = append(diagnostics, catalog.checkDefaults(paths.App, app)...)
	return catalog, diagnostics
}

func loadStorages(path string) ([]sourcedDocument[StorageProvider], Diagnostics) {
	storage, err := (YAMLFile[StorageConfig]{Path: path}).Read()
	if err != nil {
		return nil, Diagnostics{newDiagnostic("storage.load", err.Error(), path, 1)}
	}
	if len(storage) == 0 {
		return nil, Diagnostics{newDiagnostic(
			"storage.empty",
			"at least one storage provider is required",
			path,
			1,
		)}
	}

	documents := make([]sourcedDocument[StorageProvider], 0, len(storage))
	var diagnostics Diagnostics
	for index, provider := range storage {
		if err := VerifyStorageProvider(provider); err != nil {
			diagnostic := newDiagnostic("storage.invalid", err.Error(), path, 1)
			diagnostic.Field = fmt.Sprintf("[%d]", index)
			if field := validationField(err); field != "" {
				diagnostic.Field += "." + field
			}
			diagnostics = append(diagnostics, diagnostic)
			continue
		}
		documents = append(documents, sourcedDocument[StorageProvider]{
			Value:  provider,
			Source: Source{File: path, Document: 1},
		})
	}
	return documents, diagnostics
}

// indexDefinitions groups before publishing so a duplicate invalidates every
// conflicting definition instead of making discovery order significant.
func indexDefinitions[T any](
	documents []sourcedDocument[T],
	kind string,
	name func(T) string,
	diagnostics Diagnostics,
) (map[string]catalogEntry[T], Diagnostics) {
	grouped := make(map[string][]sourcedDocument[T], len(documents))
	for _, document := range documents {
		key := normalizeName(name(document.Value))
		grouped[key] = append(grouped[key], document)
	}

	index := make(map[string]catalogEntry[T], len(grouped))
	for key, definitions := range grouped {
		if len(definitions) > 1 {
			for _, definition := range definitions {
				diagnostic := newDiagnostic(
					kind+".duplicate",
					fmt.Sprintf("duplicate %s name %q", kind, name(definition.Value)),
					definition.Source.File,
					definition.Source.Document,
				)
				diagnostic.Field = "name"
				diagnostics = append(diagnostics, diagnostic)
			}
			continue
		}
		index[key] = catalogEntry[T]{
			value:  definitions[0].Value,
			source: definitions[0].Source,
		}
	}
	return index, diagnostics
}

func (c *Catalog) checkDefaults(path string, app AppConfig) Diagnostics {
	var diagnostics Diagnostics
	if _, ok := c.storages[normalizeName(app.DefaultStorageName)]; !ok {
		diagnostic := newDiagnostic(
			"application.default_storage_missing",
			fmt.Sprintf("default storage %q does not exist", app.DefaultStorageName),
			path,
			1,
		)
		diagnostic.Field = "default_storage_name"
		diagnostics = append(diagnostics, diagnostic)
	}
	if _, ok := c.collections[normalizeName(app.DefaultCollection)]; !ok {
		diagnostic := newDiagnostic(
			"application.default_collection_missing",
			fmt.Sprintf("default collection %q does not exist", app.DefaultCollection),
			path,
			1,
		)
		diagnostic.Field = "default_collection"
		diagnostics = append(diagnostics, diagnostic)
	}
	return diagnostics
}
