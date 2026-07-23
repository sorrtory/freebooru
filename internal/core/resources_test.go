package core

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestCollectionResourcesListAndImportExplicitTag(t *testing.T) {
	app := newRealImportTestCore(t)
	if err := os.WriteFile(
		filepath.Join(app.paths.Tags, "artist.yaml"),
		[]byte("name: artist\ntype: text\ncomment: Creator\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(app.paths.Tags, "series.yaml"),
		[]byte("name: series\ntype: text\ndemand:\n  - tag: artist\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(app.paths.Tags, "creator.yaml"),
		[]byte("name: creator\ntype: text\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}

	items, err := app.ListCollectionTagInfo(t.Context(), "main")
	if err != nil {
		t.Fatalf("ListCollectionTagInfo() error = %v", err)
	}
	artist := findTagInfo(items, "artist")
	if artist == nil || artist.Imported || artist.Comment != "Creator" {
		t.Fatalf("artist = %#v", artist)
	}
	storage := findTagInfo(items, "storage")
	if storage == nil || !storage.Imported || !storage.Required || len(storage.Groups) != 1 || storage.Groups[0] != "storage" {
		t.Fatalf("storage = %#v", storage)
	}
	if err := app.ImportCollectionTag(t.Context(), "main", "series"); err == nil {
		t.Fatal("invalid relationship ImportCollectionTag() error = nil")
	}
	storedAfterRollback, err := (config.YAMLFile[config.CollectionConfig]{
		Path: filepath.Join(app.paths.Collections, "main.yaml"),
	}).Read()
	if err != nil || hasTagReference(storedAfterRollback.Tags.Import, "series") {
		t.Fatalf("rollback collection = %#v, %v", storedAfterRollback, err)
	}
	if err := app.ImportCollectionTag(t.Context(), "main", "artist"); err != nil {
		t.Fatalf("ImportCollectionTag() error = %v", err)
	}
	if err := app.ImportCollectionTag(t.Context(), "main", "artist"); err == nil {
		t.Fatal("duplicate ImportCollectionTag() error = nil")
	}
	var wait sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results <- app.ImportCollectionTag(t.Context(), "main", "creator")
		}()
	}
	wait.Wait()
	close(results)
	successes := 0
	for result := range results {
		if result == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent import successes = %d, want 1", successes)
	}

	collectionConfig, _, ok := app.catalog.Collection("main")
	if !ok || !hasTagReference(collectionConfig.Tags.Import, "artist") {
		t.Fatalf("collection = %#v", collectionConfig)
	}
	stored, err := (config.YAMLFile[config.CollectionConfig]{
		Path: filepath.Join(app.paths.Collections, "main.yaml"),
	}).Read()
	if err != nil || !hasTagReference(stored.Tags.Import, "artist") {
		t.Fatalf("stored collection = %#v, %v", stored, err)
	}
}

func TestCollectionTagInfoIncludesPersistedTextValueHints(t *testing.T) {
	app := newRealImportTestCore(t)
	source := filepath.Join(t.TempDir(), "hint.txt")
	if err := os.WriteFile(source, []byte("hint"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Import(t.Context(), ImportRequest{Collection: "main", SourcePath: source, Tags: map[string]any{"rating": "safe", "title": "Studio Trigger"}}); err != nil {
		t.Fatal(err)
	}
	items, err := app.ListCollectionTagInfo(t.Context(), "main")
	if err != nil {
		t.Fatal(err)
	}
	title := findTagInfo(items, "title")
	if title == nil || !hasPredefinedValue(title.Values, "Studio Trigger") {
		t.Fatalf("title hints = %#v", title)
	}
}

func hasPredefinedValue(values []config.PredefinedValue, want string) bool {
	for _, value := range values {
		if value.Val == want {
			return true
		}
	}
	return false
}

func findTagInfo(items []CollectionTagInfo, name string) *CollectionTagInfo {
	for index := range items {
		if items[index].Name == name {
			return &items[index]
		}
	}
	return nil
}

func hasTagReference(references []config.TagReference, name string) bool {
	for _, reference := range references {
		if reference.Tag == name {
			return true
		}
	}
	return false
}
