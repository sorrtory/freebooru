package config

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestCatalogBuildsGroupAndValueIndexes(t *testing.T) {
	paths := writeCatalogFixture(t)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	group, ok := catalog.Group("CONTENT")
	if !ok || group.Name != "content" {
		t.Fatalf("Group(CONTENT) = %#v, %t", group, ok)
	}
	tags := catalog.TagsInGroup("content")
	if len(tags) != 2 || tags[0].Name != "artist" || tags[1].Name != "rating" {
		t.Fatalf("TagsInGroup(content) = %#v", tags)
	}
	values := catalog.DeclaredValues("RATING")
	if len(values) != 1 || values[0].Val != "safe" {
		t.Fatalf("DeclaredValues(RATING) = %#v", values)
	}
	values[0].Demand[0].Tag = "changed"
	again := catalog.DeclaredValues("rating")
	if again[0].Demand[0].Tag != "reviewed" {
		t.Fatalf("DeclaredValues() exposed catalog values: %#v", again)
	}
}

func TestCatalogSearchTagsUsesNormalizedPrefix(t *testing.T) {
	paths := writeCatalogFixture(t)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	tags := catalog.SearchTags("AR")
	if len(tags) != 1 || tags[0].Name != "artist" {
		t.Fatalf("SearchTags(AR) = %#v", tags)
	}
}

func TestLoadCatalogExpandsCollectionGroups(t *testing.T) {
	paths := writeCatalogFixture(t)
	content := "name: main\ntags:\n  require:\n    - storage: DEFAULT\n  import:\n    - group: CONTENT\n    - tag: Rating\n"
	writeTestFile(t, filepath.Join(paths.Collections, "main.yaml"), content)

	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	references, ok := catalog.CollectionReferences("main")
	if !ok {
		t.Fatal("CollectionReferences(main) not found")
	}
	want := []TagReference{
		{Storage: "default"},
		{Tag: "artist"},
		{Tag: "rating"},
	}
	if !reflect.DeepEqual(references.Imported, want) {
		t.Fatalf("Imported = %#v, want %#v", references.Imported, want)
	}
}

func TestLoadCatalogExcludesCollectionsWithMissingTagOrGroup(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		code      string
	}{
		{name: "tag", reference: "tag: missing", code: "collection.tag_missing"},
		{name: "group", reference: "group: missing", code: "collection.group_missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := writeCatalogFixture(t)
			content := "name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - " + tt.reference + "\n"
			writeTestFile(t, filepath.Join(paths.Collections, "main.yaml"), content)
			catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
			if _, _, ok := catalog.Collection("main"); ok {
				t.Fatal("Collection(main) exists with an unresolved reference")
			}
			if !diagnosticsContainCode(diagnostics, tt.code) {
				t.Fatalf("LoadCatalog() diagnostics = %#v, want %q", diagnostics, tt.code)
			}
		})
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
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), "name: artist\ntype: text\ngroups: [content]\n")
	writeTestFile(
		t,
		filepath.Join(paths.Tags, "rating.yaml"),
		"name: rating\ntype: value\ngroups: [content]\nvalues:\n  - val: safe\n    demand:\n      - tag: reviewed\n",
	)
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
