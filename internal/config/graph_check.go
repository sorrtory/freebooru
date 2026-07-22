package config

import (
	"fmt"
	"sort"
)

// CheckGraph validates facts that depend on both the relationship graph and
// the catalog. It reports every discoverable edge error without modifying the
// catalog, graph, or collection imports.
func CheckGraph(catalog *Catalog, graph *Graph) Diagnostics {
	var diagnostics Diagnostics
	for _, edge := range graph.edges {
		diagnostics = append(diagnostics, checkEdgeTarget(catalog, edge)...)
		diagnostics = append(diagnostics, checkSelfRelationship(edge)...)
	}
	return diagnostics
}

func checkEdgeTarget(catalog *Catalog, edge Edge) Diagnostics {
	if _, ok := catalog.tags[normalizeName(edge.TargetTag)]; !ok {
		return Diagnostics{edgeDiagnostic(
			edge,
			"relationship.target_missing",
			fmt.Sprintf(
				"relationship from %q targets missing tag %q",
				edge.Source.Tag,
				edge.TargetTag,
			),
		)}
	}

	var diagnostics Diagnostics
	for _, key := range sortedCollectionKeys(catalog.collections) {
		if !collectionImportsTag(catalog.references[key], edge.Source.Tag) {
			continue
		}
		if collectionImportsTag(catalog.references[key], edge.TargetTag) {
			continue
		}
		diagnostics = append(diagnostics, edgeDiagnostic(
			edge,
			"relationship.target_unavailable",
			fmt.Sprintf(
				"collection %q imports source tag %q but not target tag %q",
				catalog.collections[key].value.Name,
				edge.Source.Tag,
				edge.TargetTag,
			),
		))
	}
	return diagnostics
}

func checkSelfRelationship(edge Edge) Diagnostics {
	if normalizeName(edge.Source.Tag) != normalizeName(edge.TargetTag) {
		return nil
	}
	switch edge.Kind {
	case RelationshipDemand:
		return Diagnostics{edgeDiagnostic(
			edge,
			"relationship.self_demand",
			fmt.Sprintf("tag %q cannot demand itself", edge.Source.Tag),
		)}
	case RelationshipConflict:
		return Diagnostics{edgeDiagnostic(
			edge,
			"relationship.self_conflict",
			fmt.Sprintf("tag %q cannot conflict with itself", edge.Source.Tag),
		)}
	default:
		return nil
	}
}

func collectionImportsTag(references ResolvedReferences, tagName string) bool {
	want := normalizeName(tagName)
	for _, reference := range references.Imported {
		if normalizeName(reference.Tag) == want {
			return true
		}
		if want == "storage" && reference.Storage != "" {
			return true
		}
	}
	return false
}

func sortedCollectionKeys(collections map[string]catalogEntry[CollectionConfig]) []string {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func edgeDiagnostic(edge Edge, code, message string) Diagnostic {
	diagnostic := newDiagnostic(
		code,
		message,
		edge.Location.File,
		edge.Location.Document,
	)
	diagnostic.Field = edge.Location.Field
	return diagnostic
}
