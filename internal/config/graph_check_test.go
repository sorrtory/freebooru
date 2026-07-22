package config

import (
	"path/filepath"
	"testing"
)

func TestCheckGraphReportsMissingTarget(t *testing.T) {
	paths := writeCatalogFixture(t)
	artist := "name: artist\ntype: text\nsuggest:\n  - tag: missing\n"
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), artist)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	graph := BuildGraph(catalog)
	diagnostics = CheckGraph(catalog, graph)
	if !diagnosticsContainCode(diagnostics, "relationship.target_missing") {
		t.Fatalf("CheckGraph() diagnostics = %#v", diagnostics)
	}
	if diagnostics[0].Field != "suggest[0]" {
		t.Fatalf("diagnostic location = %#v", diagnostics[0])
	}
}

func TestCheckGraphRequiresTargetInSourceCollections(t *testing.T) {
	paths := writeCatalogFixture(t)
	artist := "name: artist\ntype: text\ndemand:\n  - tag: reviewed\n"
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), artist)
	collection := "name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - tag: artist\n"
	writeTestFile(t, filepath.Join(paths.Collections, "main.yaml"), collection)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	graph := BuildGraph(catalog)
	diagnostics = CheckGraph(catalog, graph)
	if !diagnosticsContainCode(diagnostics, "relationship.target_unavailable") {
		t.Fatalf("CheckGraph() diagnostics = %#v", diagnostics)
	}
	references, _ := catalog.CollectionReferences("main")
	for _, reference := range references.Imported {
		if reference.Tag == "reviewed" {
			t.Fatal("CheckGraph() auto-imported relationship target")
		}
	}
}

func TestCheckGraphAcceptsTargetImportedByGroup(t *testing.T) {
	paths := writeCatalogFixture(t)
	artist := "name: artist\ntype: text\ndemand:\n  - tag: reviewed\n"
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), artist)
	collection := "name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - tag: artist\n    - group: workflow\n"
	writeTestFile(t, filepath.Join(paths.Collections, "main.yaml"), collection)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	if diagnostics := CheckGraph(catalog, BuildGraph(catalog)); diagnostics.HasErrors() {
		t.Fatalf("CheckGraph() diagnostics = %#v", diagnostics)
	}
}

func TestCheckGraphRejectsSelfDemandAndConflict(t *testing.T) {
	paths := writeCatalogFixture(t)
	artist := "name: artist\ntype: text\n" +
		"demand:\n  - tag: ARTIST\n" +
		"conflict:\n  - tag: artist\n"
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), artist)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	diagnostics = CheckGraph(catalog, BuildGraph(catalog))
	for _, code := range []string{
		"relationship.self_demand",
		"relationship.self_conflict",
	} {
		if !diagnosticsContainCode(diagnostics, code) {
			t.Errorf("CheckGraph() diagnostics = %#v, want %q", diagnostics, code)
		}
	}
}
