package core

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

func TestSetTagPersistsEveryTagType(t *testing.T) {
	tests := []struct {
		name  string
		tag   string
		value any
	}{
		{name: "bool", tag: "flag", value: true},
		{name: "text", tag: "title", value: "example"},
		{name: "int", tag: "score", value: int64(7)},
		{name: "date", tag: "day", value: "2026-07-22"},
		{name: "datetime", tag: "instant", value: "2026-07-22T10:30:00Z"},
		{name: "value", tag: "rating", value: "questionable"},
		{name: "multivalue", tag: "labels", value: []string{"first", "second"}},
	}
	for index, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := strings.Repeat(string(rune('1'+index)), 64)
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

			if _, err := app.SetTag(t.Context(), TagMutationRequest{
				SHA256: hash,
				Tag:    tt.tag,
				Value:  tt.value,
			}); err != nil {
				t.Fatalf("SetTag() error = %v", err)
			}
			if database.setTag == nil {
				t.Fatal("SetTag() did not persist a record")
			}
			got, err := persistedTagValue(*database.setTag)
			if err != nil {
				t.Fatalf("persistedTagValue() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.value) {
				t.Fatalf("persisted value = %#v, want %#v", got, tt.value)
			}
		})
	}
}

func TestSetTagReturnsGraphViolationsSuggestionsAndReasons(t *testing.T) {
	hash := strings.Repeat("8", 64)
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
		Tag:    "trigger",
		Value:  true,
	})
	if err == nil {
		t.Fatal("SetTag() error = nil, want graph validation error")
	}
	if len(result.Evaluation.MissingDemands) != 1 ||
		result.Evaluation.MissingDemands[0].Reason != "triggered files need a title" {
		t.Fatalf("missing demands = %#v", result.Evaluation.MissingDemands)
	}
	if len(result.Evaluation.ActiveConflicts) != 1 ||
		result.Evaluation.ActiveConflicts[0].Reason != "trigger conflicts with reviewed" {
		t.Fatalf("active conflicts = %#v", result.Evaluation.ActiveConflicts)
	}
	if len(result.Evaluation.Suggestions) != 1 ||
		result.Evaluation.Suggestions[0].Reason != "a score would help" {
		t.Fatalf("suggestions = %#v", result.Evaluation.Suggestions)
	}
	if database.setTag != nil {
		t.Fatalf("invalid mutation persisted = %#v", database.setTag)
	}
}

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
