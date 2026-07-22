package core

import (
	"errors"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

func TestAddTagValidatesThenPersistsNewAssignment(t *testing.T) {
	hash := strings.Repeat("e", 64)
	rating := "safe"
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: hash,
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
		},
		Storages: []string{"default"},
	}}}
	app := newImportTestCoreWithDatabase(t, database)

	result, err := app.AddTag(t.Context(), TagMutationRequest{
		SHA256: hash,
		Tag:    "score",
		Value:  int64(7),
	})
	if err != nil {
		t.Fatalf("AddTag() error = %v", err)
	}
	if !result.Changed || database.addedTag == nil || database.addedTag.IntegerValue == nil ||
		*database.addedTag.IntegerValue != 7 {
		t.Fatalf("AddTag() result = %#v, persisted = %#v", result, database.addedTag)
	}
}

func TestAddTagRejectsScalarReplacementBeforePersistence(t *testing.T) {
	hash := strings.Repeat("f", 64)
	rating := "safe"
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: hash,
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
		},
		Storages: []string{"default"},
	}}}
	app := newImportTestCoreWithDatabase(t, database)

	_, err := app.AddTag(t.Context(), TagMutationRequest{
		SHA256: hash,
		Tag:    "rating",
		Value:  "questionable",
	})
	if !errors.Is(err, collection.ErrTagAlreadyAssigned) {
		t.Fatalf("AddTag() error = %v, want ErrTagAlreadyAssigned", err)
	}
	if database.addedTag != nil {
		t.Fatalf("persisted tag = %#v", database.addedTag)
	}
}

func TestProposeTagMutationMergesMultivalueCaseInsensitively(t *testing.T) {
	values := map[string]any{"labels": []string{"first"}}
	err := proposeTagMutation(values, preparedTagAssignment{
		tag:     config.TagConfig{Name: "labels", Type: config.TagTypeMultivalue},
		value:   []string{"FIRST", "second"},
		present: true,
	}, tagMutationAdd)
	if err != nil {
		t.Fatalf("proposeTagMutation() error = %v", err)
	}
	got := values["labels"].([]string)
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("merged values = %#v", got)
	}
}

func TestSetTagValidatesThenPersistsCanonicalAssignment(t *testing.T) {
	hash := strings.Repeat("a", 64)
	rating := "safe"
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: hash,
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
		},
		Storages: []string{"default"},
	}}}
	app := newImportTestCoreWithDatabase(t, database)

	result, err := app.SetTag(t.Context(), TagMutationRequest{
		SHA256: hash,
		Tag:    "RATING",
		Value:  "QUESTIONABLE",
	})
	if err != nil {
		t.Fatalf("SetTag() error = %v", err)
	}
	if !result.Changed || database.setTag == nil || database.setTag.Name != "rating" ||
		database.setTag.TextValue == nil || *database.setTag.TextValue != "questionable" {
		t.Fatalf("SetTag() result = %#v, persisted = %#v", result, database.setTag)
	}
	if !database.closed {
		t.Fatal("SetTag() did not close database")
	}
}

func TestSetTagRejectsRequiredBooleanRemovalBeforePersistence(t *testing.T) {
	hash := strings.Repeat("b", 64)
	rating := "safe"
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: hash,
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
		},
		Storages: []string{"default"},
	}}}
	app := newImportTestCoreWithDatabase(t, database)

	_, err := app.SetTag(t.Context(), TagMutationRequest{
		SHA256: hash,
		Tag:    "reviewed",
		Value:  false,
	})
	if err == nil || !strings.Contains(err.Error(), "required tag") {
		t.Fatalf("SetTag() error = %v, want required tag error", err)
	}
	if database.removedTag != "" || database.setTag != nil {
		t.Fatalf("mutation persisted: set=%#v removed=%q", database.setTag, database.removedTag)
	}
	if !database.closed {
		t.Fatal("SetTag() did not close database after validation failure")
	}
}

func TestRemoveTagValidatesThenRemovesOptionalAssignment(t *testing.T) {
	hash := strings.Repeat("c", 64)
	rating := "safe"
	score := int64(7)
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: hash,
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
			{Name: "score", Type: "int", IntegerValue: &score},
		},
		Storages: []string{"default"},
	}}}
	app := newImportTestCoreWithDatabase(t, database)

	result, err := app.RemoveTag(t.Context(), TagRemovalRequest{
		SHA256: hash,
		Tag:    "SCORE",
	})
	if err != nil {
		t.Fatalf("RemoveTag() error = %v", err)
	}
	if !result.Changed || database.removedTag != "score" {
		t.Fatalf("RemoveTag() result = %#v, removed = %q", result, database.removedTag)
	}
	if !database.closed {
		t.Fatal("RemoveTag() did not close database")
	}
}

func TestRemoveTagRejectsRequiredAssignmentBeforePersistence(t *testing.T) {
	hash := strings.Repeat("d", 64)
	rating := "safe"
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: hash,
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
		},
		Storages: []string{"default"},
	}}}
	app := newImportTestCoreWithDatabase(t, database)

	_, err := app.RemoveTag(t.Context(), TagRemovalRequest{
		SHA256: hash,
		Tag:    "rating",
	})
	if err == nil || !strings.Contains(err.Error(), "required tag") {
		t.Fatalf("RemoveTag() error = %v, want required tag error", err)
	}
	if database.removedTag != "" {
		t.Fatalf("removed required tag %q", database.removedTag)
	}
}
