package core

import (
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
)

func TestSearchCollectionTagsReturnsOnlyImportedPrefixMatches(t *testing.T) {
	app := newImportTestCore(t)
	tags, err := app.SearchCollectionTags("", "RA")
	if err != nil {
		t.Fatalf("SearchCollectionTags() error = %v", err)
	}
	if len(tags) != 1 || tags[0].Name != "rating" {
		t.Fatalf("SearchCollectionTags() = %#v", tags)
	}
	storageTags, err := app.SearchCollectionTags("main", "STOR")
	if err != nil {
		t.Fatalf("SearchCollectionTags(storage) error = %v", err)
	}
	if len(storageTags) != 1 || storageTags[0].Name != "storage" {
		t.Fatalf("SearchCollectionTags(storage) = %#v", storageTags)
	}
}

func TestAllowedValuesUsesPersistedFileStateAndSessionEvaluator(t *testing.T) {
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

	values, err := app.AllowedValues(t.Context(), "", hash, "RATING")
	if err != nil {
		t.Fatalf("AllowedValues() error = %v", err)
	}
	if len(values) != 2 || values[0].Value.Val != "safe" || values[1].Value.Val != "questionable" {
		t.Fatalf("AllowedValues() = %#v", values)
	}
	if !database.closed {
		t.Fatal("transient hint session was not closed")
	}
}

func TestAllowedValuesRejectsUnavailableTagBeforeFileLookup(t *testing.T) {
	database := &fakeDatabase{}
	app := newImportTestCoreWithDatabase(t, database)
	_, err := app.AllowedValues(t.Context(), "main", strings.Repeat("b", 64), "missing")
	if err == nil || !strings.Contains(err.Error(), "not imported") {
		t.Fatalf("AllowedValues() error = %v", err)
	}
}
