package core

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseTagAssignmentsResolvesTypedAndRepeatedValues(t *testing.T) {
	app := newImportTestCore(t)
	values, err := app.ParseTagAssignments("", []string{
		"reviewed",
		"score:12",
		"rating:SAFE",
		"labels:first",
		"labels:second",
		"storage:default",
	})
	if err != nil {
		t.Fatalf("ParseTagAssignments() error = %v", err)
	}
	want := map[string]any{
		"reviewed": true,
		"score":    int64(12),
		"rating":   "SAFE",
		"labels":   []string{"first", "second"},
		"storage":  []string{"default"},
	}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("ParseTagAssignments() = %#v, want %#v", values, want)
	}
}

func TestParseTagAssignmentsRejectsInvalidShapeAndDuplicateScalar(t *testing.T) {
	app := newImportTestCore(t)
	for _, assignments := range [][]string{
		{"reviewed:true"},
		{"score"},
		{"score:nope"},
		{"rating:safe", "RATING:questionable"},
		{"storage"},
		{"missing:value"},
	} {
		if _, err := app.ParseTagAssignments("main", assignments); err == nil {
			t.Fatalf("ParseTagAssignments(%q) error = nil", strings.Join(assignments, ","))
		}
	}
}
