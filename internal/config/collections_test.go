package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateStarterCollectionPublishesValidatedFileWithoutOverwrite(t *testing.T) {
	paths, err := PathsFromDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	app := DefaultAppConfig()
	collection, err := CreateStarterCollection(paths, app, "pictures")
	if err != nil {
		t.Fatalf("CreateStarterCollection() error = %v", err)
	}
	if collection.Name != "pictures" || collection.Tags.Require[0].Storage != "default" {
		t.Fatalf("collection = %#v", collection)
	}
	path := filepath.Join(paths.Collections, "pictures.yaml")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
	if _, err := CreateStarterCollection(paths, app, "pictures"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate error = %v", err)
	}
}

func TestCreateStarterCollectionRejectsUnsafeNameBeforeWriting(t *testing.T) {
	paths, err := PathsFromDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CreateStarterCollection(paths, DefaultAppConfig(), "../escape"); err == nil {
		t.Fatal("CreateStarterCollection() error = nil")
	}
	if _, err := os.Stat(filepath.Join(paths.Dir, "escape.yaml")); !os.IsNotExist(err) {
		t.Fatalf("unsafe target exists: %v", err)
	}
}
