package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCatalogIndexesDefinitionsCaseInsensitively(t *testing.T) {
	paths := writeCatalogFixture(t)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	if storage, _, ok := catalog.Storage("DEFAULT"); !ok || storage.Name != "default" {
		t.Fatalf("Storage(DEFAULT) = %#v, %t", storage, ok)
	}
	if collection, _, ok := catalog.Collection("MAIN"); !ok || collection.Name != "main" {
		t.Fatalf("Collection(MAIN) = %#v, %t", collection, ok)
	}
	if tag, source, ok := catalog.Tag("ARTIST"); !ok || tag.Name != "artist" || source.Document != 1 {
		t.Fatalf("Tag(ARTIST) = %#v, %#v, %t", tag, source, ok)
	}
}

func TestLoadCatalogInvalidatesEveryDuplicate(t *testing.T) {
	paths := writeCatalogFixture(t)
	content := "name: artist\ntype: text\n---\nname: ARTIST\ntype: text\n"
	writeTestFile(t, filepath.Join(paths.Tags, "duplicate.yaml"), content)

	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if _, _, ok := catalog.Tag("artist"); ok {
		t.Fatal("Tag(artist) exists after duplicate definitions")
	}
	if got := countDiagnostics(diagnostics, "tag.duplicate"); got != 3 {
		t.Fatalf("tag.duplicate count = %d, want 3", got)
	}
}

func TestLoadCatalogReportsMissingDefaults(t *testing.T) {
	paths := writeCatalogFixture(t)
	app := DefaultAppConfig()
	app.DefaultStorageName = "archive"
	app.DefaultCollection = "missing"

	_, diagnostics := LoadCatalog(paths, app)
	for _, code := range []string{
		"application.default_storage_missing",
		"application.default_collection_missing",
	} {
		if !diagnosticsContainCode(diagnostics, code) {
			t.Errorf("LoadCatalog() diagnostics = %#v, want %q", diagnostics, code)
		}
	}
}

func TestCatalogCollectionReturnsDefensiveCopy(t *testing.T) {
	paths := writeCatalogFixture(t)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	collection, _, _ := catalog.Collection("main")
	collection.Tags.Require[0].Storage = "changed"
	again, _, _ := catalog.Collection("main")
	if again.Tags.Require[0].Storage != "default" {
		t.Fatalf("Collection() exposed catalog storage: %#v", again)
	}
}

func TestLoadCatalogResolvesRequiredStorageAsImported(t *testing.T) {
	paths := writeCatalogFixture(t)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	references, ok := catalog.CollectionReferences("MAIN")
	if !ok {
		t.Fatal("CollectionReferences(MAIN) not found")
	}
	if len(references.Required) != 1 || len(references.Imported) != 1 {
		t.Fatalf("CollectionReferences(MAIN) = %#v", references)
	}
	if references.Imported[0].Storage != "default" {
		t.Fatalf("imported storage = %q, want default", references.Imported[0].Storage)
	}
	references.Imported[0].Storage = "changed"
	again, _ := catalog.CollectionReferences("main")
	if again.Imported[0].Storage != "default" {
		t.Fatalf("CollectionReferences() exposed catalog storage: %#v", again)
	}
}

func TestLoadCatalogExcludesCollectionWithMissingStorage(t *testing.T) {
	paths := writeCatalogFixture(t)
	content := "name: main\ntags:\n  require:\n    - storage: missing\n"
	writeTestFile(t, filepath.Join(paths.Collections, "main.yaml"), content)

	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if _, _, ok := catalog.Collection("main"); ok {
		t.Fatal("Collection(main) exists with an unresolved storage")
	}
	for _, code := range []string{
		"collection.storage_missing",
		"application.default_collection_missing",
	} {
		if !diagnosticsContainCode(diagnostics, code) {
			t.Errorf("LoadCatalog() diagnostics = %#v, want %q", diagnostics, code)
		}
	}
}

func writeCatalogFixture(t *testing.T) Paths {
	t.Helper()
	paths, err := PathsFromDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Tags, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Collections, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, paths.Storage, "- name: default\n  type: local\n  path: /tmp/storage\n")
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), "name: artist\ntype: text\n")
	writeTestFile(t, filepath.Join(paths.Collections, "main.yaml"), "name: main\ntags:\n  require:\n    - storage: default\n")
	return paths
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func countDiagnostics(diagnostics Diagnostics, code string) int {
	count := 0
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			count++
		}
	}
	return count
}
