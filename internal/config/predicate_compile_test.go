package config

import (
	"strings"
	"testing"
	"time"
)

func TestCompilePredicateAcceptsSupportedPredicates(t *testing.T) {
	minimum, maximum := int64(1), int64(10)
	tests := []struct {
		name   string
		target TagConfig
		raw    Relationship
		check  func(Predicate) bool
	}{
		{
			name:   "presence",
			target: TagConfig{Name: "reviewed", Type: TagTypeBool},
			check:  func(got Predicate) bool { return got.Presence },
		},
		{
			name:   "bool false",
			target: TagConfig{Name: "reviewed", Type: TagTypeBool},
			raw:    Relationship{Is: false},
			check:  func(got Predicate) bool { value, ok := got.Is.(bool); return ok && !value },
		},
		{
			name:   "integer range",
			target: TagConfig{Name: "score", Type: TagTypeInt},
			raw:    Relationship{Min: &minimum, Max: &maximum},
			check:  func(got Predicate) bool { return *got.Min == 1 && *got.Max == 10 },
		},
		{
			name:   "time range",
			target: TagConfig{Name: "published", Type: TagTypeDate},
			raw:    Relationship{After: "2026-01-01", Before: "2026-12-31"},
			check:  func(got Predicate) bool { return got.After.Year() == 2026 && got.Before.Year() == 2026 },
		},
		{
			name: "has and not",
			target: TagConfig{
				Name:   "rating",
				Type:   TagTypeMultivalue,
				Values: []PredefinedValue{{Val: "safe"}, {Val: "explicit"}},
			},
			raw:   Relationship{Has: []any{"SAFE"}, Not: []any{"explicit"}},
			check: func(got Predicate) bool { return got.Has[0] == "safe" && got.Not[0] == "explicit" },
		},
		{
			name:   "regex",
			target: TagConfig{Name: "caption", Type: TagTypeText},
			raw:    Relationship{Regex: "foo.*bar"},
			check:  func(got Predicate) bool { return got.Regex.MatchString("foo  bar") },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compilePredicate(tt.target, tt.raw)
			if err != nil {
				t.Fatalf("compilePredicate() error = %v", err)
			}
			if !tt.check(got) {
				t.Fatalf("compilePredicate() = %#v", got)
			}
		})
	}
}

func TestCompilePredicateRejectsInvalidPredicates(t *testing.T) {
	negative, lower, upper := int64(-1), int64(10), int64(1)
	tests := []struct {
		name   string
		target TagConfig
		raw    Relationship
		want   string
	}{
		{name: "is combination", target: valueTag(TagTypeValue), raw: Relationship{Is: "safe", Not: []any{"safe"}}, want: "cannot be combined"},
		{name: "has text", target: TagConfig{Name: "caption", Type: TagTypeText}, raw: Relationship{Has: []any{"safe"}}, want: "require a value"},
		{name: "has value", target: valueTag(TagTypeValue), raw: Relationship{Has: []any{"safe"}}, want: "multivalue"},
		{name: "unknown predefined", target: valueTag(TagTypeValue), raw: Relationship{Is: "missing"}, want: "does not declare"},
		{name: "negative bound", target: TagConfig{Name: "score", Type: TagTypeInt}, raw: Relationship{Min: &negative}, want: "at least 0"},
		{name: "reversed integer range", target: TagConfig{Name: "score", Type: TagTypeInt}, raw: Relationship{Min: &lower, Max: &upper}, want: "must not exceed"},
		{name: "invalid date", target: TagConfig{Name: "published", Type: TagTypeDate}, raw: Relationship{Before: "2026-02-30"}, want: "must match"},
		{name: "reversed date range", target: TagConfig{Name: "published", Type: TagTypeDate}, raw: Relationship{After: "2026-12-31", Before: "2026-01-01"}, want: "earlier"},
		{name: "invalid regex", target: TagConfig{Name: "caption", Type: TagTypeText}, raw: Relationship{Regex: "["}, want: "compile regex"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compilePredicate(tt.target, tt.raw)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("compilePredicate() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestCompilePredicateParsesRFC3339Is(t *testing.T) {
	predicate, err := compilePredicate(
		TagConfig{Name: "created_at", Type: TagTypeDatetime},
		Relationship{Is: "2026-07-22T12:30:00+03:00"},
	)
	if err != nil {
		t.Fatalf("compilePredicate() error = %v", err)
	}
	got, ok := predicate.Is.(time.Time)
	if !ok || got.Year() != 2026 {
		t.Fatalf("compiled is = %#v", predicate.Is)
	}
}
