package config

import (
	"fmt"
	"sort"
)

type relationshipList struct {
	kind          RelationshipKind
	field         string
	relationships []Relationship
}

// BuildGraph indexes relationships from valid catalog tags. Target validation
// and predicate compilation are separate Phase 10 operations.
func BuildGraph(catalog *Catalog) *Graph {
	graph := &Graph{
		outgoing: make(map[string][]Edge),
		incoming: make(map[string][]Edge),
	}
	tagKeys := make([]string, 0, len(catalog.tags))
	for key := range catalog.tags {
		tagKeys = append(tagKeys, key)
	}
	sort.Strings(tagKeys)

	for _, key := range tagKeys {
		entry := catalog.tags[key]
		graph.addRelationshipLists(
			SourceCondition{Tag: entry.value.Name},
			entry.source,
			"",
			tagRelationshipLists(entry.value),
		)
		for valueIndex, value := range entry.value.Values {
			graph.addRelationshipLists(
				SourceCondition{Tag: entry.value.Name, Value: value.Val},
				entry.source,
				fmt.Sprintf("values[%d].", valueIndex),
				valueRelationshipLists(value),
			)
		}
	}
	return graph
}

func tagRelationshipLists(tag TagConfig) []relationshipList {
	return relationshipLists(tag.Suggest, tag.Demand, tag.Conflict)
}

func valueRelationshipLists(value PredefinedValue) []relationshipList {
	return relationshipLists(value.Suggest, value.Demand, value.Conflict)
}

func relationshipLists(
	suggest []Relationship,
	demand []Relationship,
	conflict []Relationship,
) []relationshipList {
	return []relationshipList{
		{kind: RelationshipSuggest, field: "suggest", relationships: suggest},
		{kind: RelationshipDemand, field: "demand", relationships: demand},
		{kind: RelationshipConflict, field: "conflict", relationships: conflict},
	}
}

func (g *Graph) addRelationshipLists(
	source SourceCondition,
	location Source,
	fieldPrefix string,
	lists []relationshipList,
) {
	for _, list := range lists {
		for index, relationship := range list.relationships {
			edge := Edge{
				Kind:      list.kind,
				Source:    source,
				TargetTag: relationship.Tag,
				Predicate: predicateFromRelationship(relationship),
				Reason:    relationship.Reason,
				Location: Location{
					Source: location,
					Field:  fmt.Sprintf("%s%s[%d]", fieldPrefix, list.field, index),
				},
			}
			g.addEdge(edge)
		}
	}
}

func (g *Graph) addEdge(edge Edge) {
	sourceKey := sourceConditionKey(edge.Source)
	targetKey := normalizeName(edge.TargetTag)
	g.outgoing[sourceKey] = append(g.outgoing[sourceKey], edge)
	g.incoming[targetKey] = append(g.incoming[targetKey], edge)
}

func predicateFromRelationship(relationship Relationship) Predicate {
	return Predicate{
		Has:    append([]any(nil), relationship.Has...),
		Is:     relationship.Is,
		Not:    append([]any(nil), relationship.Not...),
		Min:    relationship.Min,
		Max:    relationship.Max,
		Before: relationship.Before,
		After:  relationship.After,
		Regex:  relationship.Regex,
	}
}

func sourceConditionKey(source SourceCondition) string {
	return normalizeName(source.Tag) + "\x00" + normalizeName(source.Value)
}
