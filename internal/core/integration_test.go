package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestConfigIntegrationKeepsIndependentCollectionsUsable(t *testing.T) {
	paths, appConfig := provisionTestConfig(t)
	home := filepath.Dir(paths.Dir)
	storages := config.StorageConfig{
		{Name: "default", Type: "local", Path: appConfig.DefaultStoragePath},
		{Name: "archive", Type: "local", Path: filepath.Join(home, "storage", "archive")},
	}
	if err := (config.YAMLFile[config.StorageConfig]{Path: paths.Storage}).Write(storages); err != nil {
		t.Fatal(err)
	}
	writeIntegrationFile(
		t,
		filepath.Join(paths.Tags, "people", "artist.yaml"),
		"name: artist\ntype: text\ngroups: [content]\ndemand:\n  - tag: rating\n    is: safe\n",
	)
	writeIntegrationFile(
		t,
		filepath.Join(paths.Tags, "rating.yaml"),
		"name: rating\ntype: value\ngroups: [content]\nvalues:\n  - val: safe\n  - val: explicit\n---\nname: broken_tag\ntype: bool\ndemand:\n  - tag: missing\n",
	)
	writeCollectionFixture(t, filepath.Join(paths.Collections, "main.yaml"), "main", "default")
	writeCollectionFixture(t, filepath.Join(paths.Collections, "archive.yaml"), "archive", "archive")
	writeCollectionFixture(t, filepath.Join(paths.Collections, "broken.yaml"), "broken", "missing")

	var opened []string
	app, err := New(testLogger(), paths, func(_ context.Context, path string) (CollectionDatabase, error) {
		opened = append(opened, path)
		return &fakeDatabase{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	diagnostics := app.CheckConfig(t.Context())
	for _, code := range []string{
		"collection.storage_missing",
		"relationship.target_missing",
	} {
		if !hasDiagnosticCode(diagnostics, code) {
			t.Errorf("CheckConfig() diagnostics = %#v, want %q", diagnostics, code)
		}
	}
	for _, name := range []string{"main", "archive"} {
		if err := app.OpenCollection(t.Context(), name); err != nil {
			t.Fatalf("OpenCollection(%s) error = %v", name, err)
		}
		if err := app.CloseCollection(); err != nil {
			t.Fatalf("CloseCollection(%s) error = %v", name, err)
		}
	}
	if err := app.OpenCollection(t.Context(), "broken"); err == nil {
		t.Fatal("OpenCollection(broken) error = nil")
	}
	if len(opened) != 2 {
		t.Fatalf("opened paths = %#v", opened)
	}
	tags, err := app.SearchTags("ART")
	if err != nil || len(tags) != 1 || tags[0].Name != "artist" {
		t.Fatalf("SearchTags(ART) = %#v, %v", tags, err)
	}
}

func writeCollectionFixture(t *testing.T, path, name, storage string) {
	t.Helper()
	collection := config.CollectionConfig{
		Name: name,
		Tags: config.CollectionTagImports{
			Require: []config.TagReference{{Storage: storage}},
			Import:  []config.TagReference{{Group: "content"}},
		},
	}
	if err := (config.YAMLFile[config.CollectionConfig]{
		Path:     path,
		Validate: config.VerifyCollectionConfig,
	}).Write(collection); err != nil {
		t.Fatal(err)
	}
}

func writeIntegrationFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
