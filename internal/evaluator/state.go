// Package evaluator checks file tag state against a compiled config graph.
package evaluator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sorrtory/freebooru/internal/config"
)

type assignment struct {
	tag   config.TagConfig
	value any
}

// FileState is an immutable snapshot of tags assigned to one indexed file.
// Boolean false represents absence and is therefore not stored as an active tag.
type FileState struct {
	assignments map[string]assignment
}

// NewFileState validates assignments against the catalog and copies mutable
// values before publishing the state.
func NewFileState(catalog *config.Catalog, values map[string]any) (FileState, error) {
	state := FileState{assignments: make(map[string]assignment, len(values))}
	seen := make(map[string]struct{}, len(values))
	for name, value := range values {
		key := normalizeName(name)
		if _, duplicate := seen[key]; duplicate {
			return FileState{}, fmt.Errorf("tag %q is assigned more than once", name)
		}
		seen[key] = struct{}{}
		tag, _, ok := catalog.Tag(name)
		if !ok {
			return FileState{}, fmt.Errorf("tag %q does not exist", name)
		}
		if err := config.VerifyTagValue(tag, value); err != nil {
			return FileState{}, err
		}
		if tag.Type == config.TagTypeBool && !value.(bool) {
			continue
		}
		state.assignments[key] = assignment{tag: tag, value: cloneValue(value)}
	}
	return state, nil
}

// Value returns a defensive copy of one assigned value.
func (s FileState) Value(tagName string) (any, bool) {
	assigned, ok := s.assignments[normalizeName(tagName)]
	if !ok {
		return nil, false
	}
	return cloneValue(assigned.value), true
}

// ActiveSources returns graph source conditions activated by this state.
// Every assigned tag activates its presence condition; predefined values also
// activate their value-specific conditions.
func (s FileState) ActiveSources() []config.SourceCondition {
	keys := make([]string, 0, len(s.assignments))
	for key := range s.assignments {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sources []config.SourceCondition
	for _, key := range keys {
		assigned := s.assignments[key]
		sources = append(sources, config.SourceCondition{Tag: assigned.tag.Name})
		sources = append(sources, valueSources(assigned)...)
	}
	return sources
}

func valueSources(assigned assignment) []config.SourceCondition {
	switch assigned.tag.Type {
	case config.TagTypeValue:
		return []config.SourceCondition{{
			Tag:   assigned.tag.Name,
			Value: assigned.value.(string),
		}}
	case config.TagTypeMultivalue:
		values := append([]string(nil), assigned.value.([]string)...)
		sort.Slice(values, func(i, j int) bool {
			return normalizeName(values[i]) < normalizeName(values[j])
		})
		sources := make([]config.SourceCondition, 0, len(values))
		for _, value := range values {
			sources = append(sources, config.SourceCondition{
				Tag:   assigned.tag.Name,
				Value: value,
			})
		}
		return sources
	default:
		return nil
	}
}

func cloneValue(value any) any {
	if values, ok := value.([]string); ok {
		return append([]string(nil), values...)
	}
	return value
}

func normalizeName(name string) string {
	return strings.ToLower(name)
}
