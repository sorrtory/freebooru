package config

// Outgoing returns all edges activated by a source condition.
func (g *Graph) Outgoing(source SourceCondition) []Edge {
	return cloneEdges(g.outgoing[sourceConditionKey(source)])
}

// Incoming returns all edges that target a tag, also known as backlinks.
func (g *Graph) Incoming(targetTag string) []Edge {
	return cloneEdges(g.incoming[normalizeName(targetTag)])
}

// Suggestions returns suggestion edges activated by a source condition.
func (g *Graph) Suggestions(source SourceCondition) []Edge {
	return filterEdges(g.outgoing[sourceConditionKey(source)], RelationshipSuggest)
}

// Demands returns demand edges activated by a source condition.
func (g *Graph) Demands(source SourceCondition) []Edge {
	return filterEdges(g.outgoing[sourceConditionKey(source)], RelationshipDemand)
}

// Conflicts returns conflict edges activated by a source condition.
func (g *Graph) Conflicts(source SourceCondition) []Edge {
	return filterEdges(g.outgoing[sourceConditionKey(source)], RelationshipConflict)
}

// Backlinks is the editor-facing name for Incoming.
func (g *Graph) Backlinks(targetTag string) []Edge {
	return g.Incoming(targetTag)
}

func filterEdges(edges []Edge, kind RelationshipKind) []Edge {
	filtered := make([]Edge, 0)
	for _, edge := range edges {
		if edge.Kind == kind {
			filtered = append(filtered, cloneEdge(edge))
		}
	}
	return filtered
}

func cloneEdges(edges []Edge) []Edge {
	cloned := make([]Edge, len(edges))
	for index, edge := range edges {
		cloned[index] = cloneEdge(edge)
	}
	return cloned
}

func cloneEdge(edge Edge) Edge {
	edge.Predicate.Has = append([]string(nil), edge.Predicate.Has...)
	edge.Predicate.Not = append([]string(nil), edge.Predicate.Not...)
	return edge
}
