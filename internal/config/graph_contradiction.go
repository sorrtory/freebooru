package config

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// checkGraphContradictions rejects an identical target condition that is both
// demanded and forbidden by the same source condition. Cycles are intentionally
// irrelevant here because relationships are directed and evaluated per edge.
func checkGraphContradictions(graph *Graph) Diagnostics {
	demands := make(map[string]Edge)
	for _, edge := range graph.edges {
		if edge.Kind == RelationshipDemand {
			demands[edgeConditionKey(edge)] = edge
		}
	}

	var diagnostics Diagnostics
	seen := make(map[string]struct{})
	for _, edge := range graph.edges {
		if edge.Kind != RelationshipConflict {
			continue
		}
		key := edgeConditionKey(edge)
		demand, ok := demands[key]
		if !ok {
			continue
		}
		if _, reported := seen[key]; reported {
			continue
		}
		seen[key] = struct{}{}
		diagnostics = append(diagnostics, edgeDiagnostic(
			edge,
			"relationship.demand_conflict",
			fmt.Sprintf(
				"source %q both demands and conflicts with target %q; demand declared at %s:%s",
				edge.Source.Tag,
				edge.TargetTag,
				demand.Location.File,
				demand.Location.Field,
			),
		))
	}
	return diagnostics
}

func edgeConditionKey(edge Edge) string {
	return sourceConditionKey(edge.Source) + "\x00" +
		normalizeName(edge.TargetTag) + "\x00" + predicateKey(edge.Predicate)
}

func predicateKey(predicate Predicate) string {
	parts := []string{fmt.Sprintf("presence=%t", predicate.Presence)}
	if predicate.Is != nil {
		parts = append(parts, "is="+typedValueKey(predicate.Is))
	}
	parts = append(parts, "has="+sortedValuesKey(predicate.Has))
	parts = append(parts, "not="+sortedValuesKey(predicate.Not))
	parts = append(parts, "min="+intPointerKey(predicate.Min))
	parts = append(parts, "max="+intPointerKey(predicate.Max))
	parts = append(parts, "before="+timePointerKey(predicate.Before))
	parts = append(parts, "after="+timePointerKey(predicate.After))
	if predicate.Regex != nil {
		parts = append(parts, "regex="+predicate.Regex.String())
	}
	return strings.Join(parts, "\x00")
}

func typedValueKey(value any) string {
	if timestamp, ok := value.(time.Time); ok {
		return "time:" + timestamp.UTC().Format(time.RFC3339Nano)
	}
	return fmt.Sprintf("%T:%v", value, value)
}

func sortedValuesKey(values []string) string {
	cloned := append([]string(nil), values...)
	for index := range cloned {
		cloned[index] = normalizeName(cloned[index])
	}
	sort.Strings(cloned)
	return strings.Join(cloned, ",")
}

func intPointerKey(value *int64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%d", *value)
}

func timePointerKey(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
