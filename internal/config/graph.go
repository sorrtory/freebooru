package config

// RelationshipKind identifies the meaning of a directed relationship edge.
type RelationshipKind string

// Supported relationship kinds.
const (
	RelationshipSuggest  RelationshipKind = "suggest"
	RelationshipDemand   RelationshipKind = "demand"
	RelationshipConflict RelationshipKind = "conflict"
)

// SourceCondition activates on tag presence or on one predefined value.
// Value is empty for a tag-level relationship.
type SourceCondition struct {
	Tag   string
	Value string
}

// Predicate is the uncompiled target condition decoded from YAML.
// Phase 10 validates compatibility with the target tag and compiles it.
type Predicate struct {
	Has    []any
	Is     any
	Not    []any
	Min    *int64
	Max    *int64
	Before string
	After  string
	Regex  string
}

// Location identifies the exact configuration field that declared an edge.
type Location struct {
	Source
	Field string
}

// Edge is one explicit, directed configuration relationship.
type Edge struct {
	Kind      RelationshipKind
	Source    SourceCondition
	TargetTag string
	Predicate Predicate
	Reason    string
	Location  Location
}

// Graph is an immutable relationship snapshot with forward and reverse indexes.
// It contains only explicit edges; building it never infers reverse rules.
type Graph struct {
	outgoing map[string][]Edge
	incoming map[string][]Edge
}
