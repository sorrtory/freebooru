package core

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

func TestPersistedFileStateLoadsTypedAssignments(t *testing.T) {
	app := newImportTestCore(t)
	references, ok := app.catalog.CollectionReferences("main")
	if !ok {
		t.Fatal("main collection references are unavailable")
	}
	rating := "safe"
	score := int64(7)
	state, err := persistedFileState(collection.FileRecord{
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
			{Name: "score", Type: "int", IntegerValue: &score},
		},
		Storages: []string{"default"},
	}, app.catalog, newCollectionAvailability(references))
	if err != nil {
		t.Fatalf("persistedFileState() error = %v", err)
	}
	for name, want := range map[string]any{
		"reviewed": true,
		"rating":   "safe",
		"score":    int64(7),
		"storage":  []string{"default"},
	} {
		got, assigned := state.Value(name)
		if !assigned || !reflect.DeepEqual(got, want) {
			t.Errorf("Value(%q) = %#v, %t, want %#v, true", name, got, assigned, want)
		}
	}
}

func TestPersistedFileStateEvaluatesEveryRelationshipKind(t *testing.T) {
	app := newImportTestCore(t)
	references, ok := app.catalog.CollectionReferences("main")
	if !ok {
		t.Fatal("main collection references are unavailable")
	}
	rating := "safe"
	state, err := persistedFileState(collection.FileRecord{
		SHA256: strings.Repeat("a", 64),
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
			{Name: "trigger", Type: "bool"},
		},
		Storages: []string{"default"},
	}, app.catalog, newCollectionAvailability(references))
	if err != nil {
		t.Fatalf("persistedFileState() error = %v", err)
	}
	checker, err := evaluator.New(app.catalog, app.graph)
	if err != nil {
		t.Fatal(err)
	}
	result := checker.ValidateFile(state)
	if len(result.MissingDemands) != 1 ||
		result.MissingDemands[0].Reason != "triggered files need a title" {
		t.Fatalf("missing demands = %#v", result.MissingDemands)
	}
	if len(result.ActiveConflicts) != 1 ||
		result.ActiveConflicts[0].Reason != "trigger conflicts with reviewed" {
		t.Fatalf("active conflicts = %#v", result.ActiveConflicts)
	}
	if len(result.Suggestions) != 1 ||
		result.Suggestions[0].Reason != "a score would help" {
		t.Fatalf("suggestions = %#v", result.Suggestions)
	}
}

func TestOpenCollectionRejectsMissingRequiredPersistedTag(t *testing.T) {
	app := newImportTestCore(t)
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: strings.Repeat("b", 64),
		Tags:   []collection.TagRecord{{Name: "reviewed", Type: "bool"}},
		Storages: []string{
			"default",
		},
	}}}
	app.open = func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	}
	err := app.OpenCollection(t.Context(), "main")
	if err == nil || !strings.Contains(err.Error(), "required tag \"rating\"") {
		t.Fatalf("OpenCollection() error = %v", err)
	}
	if !database.closed {
		t.Fatal("invalid persisted collection database was not closed")
	}
}

func TestPersistedTagValuePreservesTypes(t *testing.T) {
	text := "example"
	number := int64(12)
	tests := []struct {
		name string
		tag  collection.TagRecord
		want any
	}{
		{name: "bool", tag: collection.TagRecord{Name: "flag", Type: "bool"}, want: true},
		{
			name: "text",
			tag:  collection.TagRecord{Name: "title", Type: "text", TextValue: &text},
			want: "example",
		},
		{
			name: "integer",
			tag:  collection.TagRecord{Name: "score", Type: "int", IntegerValue: &number},
			want: int64(12),
		},
		{
			name: "date",
			tag:  collection.TagRecord{Name: "day", Type: "date", TextValue: &text},
			want: "example",
		},
		{
			name: "datetime",
			tag:  collection.TagRecord{Name: "instant", Type: "datetime", TextValue: &text},
			want: "example",
		},
		{
			name: "value",
			tag:  collection.TagRecord{Name: "rating", Type: "value", TextValue: &text},
			want: "example",
		},
		{
			name: "multivalue",
			tag: collection.TagRecord{
				Name: "labels", Type: "multivalue", Values: []string{"one", "two"},
			},
			want: []string{"one", "two"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := persistedTagValue(tt.tag)
			if err != nil {
				t.Fatalf("persistedTagValue() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("persistedTagValue() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPersistedTagValueRejectsAmbiguousShape(t *testing.T) {
	text := "wrong"
	if _, err := persistedTagValue(collection.TagRecord{
		Name: "score", Type: "int", TextValue: &text,
	}); err == nil {
		t.Fatal("persistedTagValue() error = nil, want invalid shape error")
	}
}
