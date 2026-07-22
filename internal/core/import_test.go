package core

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestPrepareImportAppliesRequiredAndDefaultValues(t *testing.T) {
	app := newImportTestCore(t)
	prepared, err := app.prepareImport(ImportRequest{
		Tags: map[string]any{
			"RATING": "SAFE",
			"score":  int64(7),
		},
	})
	if err != nil {
		t.Fatalf("prepareImport() error = %v", err)
	}
	if prepared.collection.Name != "main" {
		t.Fatalf("collection = %q, want main", prepared.collection.Name)
	}
	if reviewed, ok := prepared.values["reviewed"].(bool); !ok || !reviewed {
		t.Fatalf("reviewed value = %#v, want required true", prepared.values["reviewed"])
	}
	if prepared.values["rating"] != "safe" {
		t.Fatalf("rating value = %#v, want canonical safe", prepared.values["rating"])
	}
	storages, ok := prepared.values["storage"].([]string)
	if !ok || len(storages) != 1 || storages[0] != "default" {
		t.Fatalf("storage value = %#v, want default", prepared.values["storage"])
	}
	if len(prepared.storages) != 1 || prepared.storages[0].Name != "default" {
		t.Fatalf("resolved storages = %#v", prepared.storages)
	}
}

func TestPrepareImportRequiresExplicitNonBooleanValue(t *testing.T) {
	app := newImportTestCore(t)
	_, err := app.prepareImport(ImportRequest{})
	if err == nil || !strings.Contains(err.Error(), "needs an explicit value") {
		t.Fatalf("prepareImport() error = %v, want required value error", err)
	}
}

func TestPrepareImportRejectsUnavailableAssignment(t *testing.T) {
	app := newImportTestCore(t)
	_, err := app.prepareImport(ImportRequest{Tags: map[string]any{
		"rating":  "safe",
		"missing": true,
	}})
	if err == nil || !strings.Contains(err.Error(), "not imported") {
		t.Fatalf("prepareImport() error = %v, want unavailable tag error", err)
	}
}

func newImportTestCore(t *testing.T) *Core {
	return newImportTestCoreWithDatabase(t, &fakeDatabase{})
}

func newImportTestCoreWithDatabase(t *testing.T, database CollectionDatabase) *Core {
	t.Helper()
	paths, _ := provisionTestConfig(t)
	tags := `name: reviewed
type: bool
---
name: rating
type: value
values:
  - val: safe
  - val: questionable
---
name: score
type: int
`
	if err := os.WriteFile(filepath.Join(paths.Tags, "import.yaml"), []byte(tags), 0o600); err != nil {
		t.Fatal(err)
	}
	collectionDocument := `name: main
tags:
  require:
    - storage: default
    - tag: reviewed
    - tag: rating
  import:
    - tag: score
`
	if err := os.WriteFile(
		filepath.Join(paths.Collections, "main.yaml"),
		[]byte(collectionDocument),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	if app.AppConfig() == (config.AppConfig{}) {
		t.Fatal("application config was not loaded")
	}
	return app
}
