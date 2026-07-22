package config

import "fmt"

// CompileGraph converts raw YAML predicates to target-tag types once. Edges
// with invalid predicates are omitted from the returned immutable graph.
// Cross-file target diagnostics remain the responsibility of CheckGraph.
func CompileGraph(catalog *Catalog, rawGraph *Graph) (*Graph, Diagnostics) {
	compiled := &Graph{
		outgoing: make(map[string][]Edge),
		incoming: make(map[string][]Edge),
	}
	var diagnostics Diagnostics
	for _, edge := range rawGraph.edges {
		target, ok := catalog.tags[normalizeName(edge.TargetTag)]
		if !ok {
			continue
		}
		predicate, err := compilePredicate(target.value, edge.raw)
		if err != nil {
			diagnostics = append(diagnostics, edgeDiagnostic(
				edge,
				"predicate.invalid",
				fmt.Sprintf(
					"relationship from %q to %q has invalid predicate: %v",
					edge.Source.Tag,
					edge.TargetTag,
					err,
				),
			))
			continue
		}
		edge.TargetTag = target.value.Name
		edge.Predicate = predicate
		compiled.addEdge(edge)
	}
	return compiled, diagnostics
}
