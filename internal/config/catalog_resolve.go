package config

import "fmt"

// resolveCollectionStorages removes collections that reference unavailable
// storage. Tag and group resolution is added when their full models exist.
func (c *Catalog) resolveCollectionStorages() Diagnostics {
	c.references = make(map[string]ResolvedReferences, len(c.collections))
	var diagnostics Diagnostics
	for key, entry := range c.collections {
		resolved, found := c.resolveReferences(entry)
		diagnostics = append(diagnostics, found...)
		if found.HasErrors() {
			delete(c.collections, key)
			continue
		}
		c.references[key] = resolved
	}
	return diagnostics
}

func (c *Catalog) resolveReferences(
	entry catalogEntry[CollectionConfig],
) (ResolvedReferences, Diagnostics) {
	required, diagnostics := c.resolveReferenceList(
		entry.value.Tags.Require,
		entry.source,
		"tags.require",
	)
	imported, found := c.resolveReferenceList(
		entry.value.Tags.Import,
		entry.source,
		"tags.import",
	)
	diagnostics = append(diagnostics, found...)
	allImported := make([]TagReference, 0, len(required)+len(imported))
	allImported = append(allImported, required...)
	allImported = append(allImported, imported...)
	return ResolvedReferences{
		Required: uniqueReferences(required),
		Imported: uniqueReferences(allImported),
	}, diagnostics
}

func (c *Catalog) resolveReferenceList(
	references []TagReference,
	source Source,
	field string,
) ([]TagReference, Diagnostics) {
	var diagnostics Diagnostics
	for index, reference := range references {
		if reference.Storage == "" {
			continue
		}
		if _, ok := c.storages[normalizeName(reference.Storage)]; ok {
			continue
		}
		diagnostic := newDiagnostic(
			"collection.storage_missing",
			fmt.Sprintf("storage %q does not exist", reference.Storage),
			source.File,
			source.Document,
		)
		diagnostic.Field = fmt.Sprintf("%s[%d].storage", field, index)
		diagnostics = append(diagnostics, diagnostic)
	}
	return references, diagnostics
}

func uniqueReferences(references []TagReference) []TagReference {
	unique := make([]TagReference, 0, len(references))
	seen := make(map[string]struct{}, len(references))
	for _, reference := range references {
		key := referenceKey(reference)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, reference)
	}
	return unique
}

func referenceKey(reference TagReference) string {
	switch {
	case reference.Tag != "":
		return "tag:" + normalizeName(reference.Tag)
	case reference.Group != "":
		return "group:" + normalizeName(reference.Group)
	default:
		return "storage:" + normalizeName(reference.Storage)
	}
}
