package config

import (
	"regexp"
	"time"
)

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

// Predicate is a target condition compiled to the target tag's concrete type.
type Predicate struct {
	Presence bool
	Has      []string
	Is       any
	Not      []string
	Min      *int64
	Max      *int64
	Before   *time.Time
	After    *time.Time
	Regex    *regexp.Regexp
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
	raw       Relationship
}

// Graph is an immutable relationship snapshot with forward and reverse indexes.
// It contains only explicit edges; building it never infers reverse rules.
type Graph struct {
	outgoing map[string][]Edge
	incoming map[string][]Edge
	edges    []Edge
}
