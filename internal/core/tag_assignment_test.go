package core

import (
	"reflect"
	"testing"
)

func TestPrepareTagAssignmentCanonicalizesAndBuildsRecord(t *testing.T) {
	app := newImportTestCore(t)
	references, ok := app.catalog.CollectionReferences("main")
	if !ok {
		t.Fatal("main collection references are unavailable")
	}

	assignment, err := app.prepareTagAssignment(
		"RATING",
		"SAFE",
		newCollectionAvailability(references),
	)
	if err != nil {
		t.Fatalf("prepareTagAssignment() error = %v", err)
	}
	if !assignment.present || assignment.tag.Name != "rating" || assignment.value != "safe" {
		t.Fatalf("prepareTagAssignment() = %#v", assignment)
	}
	if assignment.record.Name != "rating" || assignment.record.TextValue == nil ||
		*assignment.record.TextValue != "safe" {
		t.Fatalf("record = %#v", assignment.record)
	}
}

func TestPrepareTagAssignmentTreatsBooleanFalseAsAbsent(t *testing.T) {
	app := newImportTestCore(t)
	references, _ := app.catalog.CollectionReferences("main")

	assignment, err := app.prepareTagAssignment(
		"reviewed",
		false,
		newCollectionAvailability(references),
	)
	if err != nil {
		t.Fatalf("prepareTagAssignment() error = %v", err)
	}
	if assignment.present || !reflect.DeepEqual(assignment.value, false) {
		t.Fatalf("prepareTagAssignment() = %#v, want absent false", assignment)
	}
}

func TestPrepareTagAssignmentRejectsUnavailableAndInvalidValues(t *testing.T) {
	app := newImportTestCore(t)
	references, _ := app.catalog.CollectionReferences("main")
	availability := newCollectionAvailability(references)
	tests := []struct {
		name  string
		tag   string
		value any
	}{
		{name: "unavailable", tag: "missing", value: true},
		{name: "invalid type", tag: "score", value: "seven"},
		{name: "storage", tag: "storage", value: []string{"default"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := app.prepareTagAssignment(tt.tag, tt.value, availability); err == nil {
				t.Fatalf("prepareTagAssignment(%q) error = nil", tt.tag)
			}
		})
	}
}
