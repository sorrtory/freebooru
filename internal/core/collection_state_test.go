package core

import (
	"reflect"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
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
