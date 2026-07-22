package config

import (
	"path/filepath"
	"testing"
)

func TestBuildValidatedGraphRejectsDemandConflictContradiction(t *testing.T) {
	paths := writeCatalogFixture(t)
	artist := "name: artist\ntype: text\n" +
		"demand:\n  - tag: rating\n    is: SAFE\n" +
		"conflict:\n  - tag: RATING\n    is: safe\n"
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), artist)
	catalog, catalogDiagnostics := LoadCatalog(paths, DefaultAppConfig())
	if catalogDiagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", catalogDiagnostics)
	}

	graph, diagnostics := BuildValidatedGraph(catalog)
	if !diagnosticsContainCode(diagnostics, "relationship.demand_conflict") {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
	if got := graph.Demands(SourceCondition{Tag: "artist"}); len(got) != 1 {
		t.Fatalf("Demands(artist) = %#v", got)
	}
}

func TestBuildValidatedGraphAllowsDifferentTargetConditions(t *testing.T) {
	paths := writeCatalogFixture(t)
	rating := "name: rating\ntype: value\nvalues:\n  - val: safe\n  - val: explicit\n"
	writeTestFile(t, filepath.Join(paths.Tags, "rating.yaml"), rating)
	artist := "name: artist\ntype: text\n" +
		"demand:\n  - tag: rating\n    is: safe\n" +
		"conflict:\n  - tag: rating\n    is: explicit\n"
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), artist)
	catalog, catalogDiagnostics := LoadCatalog(paths, DefaultAppConfig())
	if catalogDiagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", catalogDiagnostics)
	}
	if _, diagnostics := BuildValidatedGraph(catalog); diagnostics.HasErrors() {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
}

func TestBuildValidatedGraphAllowsSuggestAndDemandCycles(t *testing.T) {
	paths := writeCatalogFixture(t)
	alpha := "name: alpha\ntype: bool\n" +
		"suggest:\n  - tag: beta\n" +
		"demand:\n  - tag: beta\n"
	beta := "name: beta\ntype: bool\n" +
		"suggest:\n  - tag: alpha\n" +
		"demand:\n  - tag: alpha\n"
	writeTestFile(t, filepath.Join(paths.Tags, "alpha.yaml"), alpha)
	writeTestFile(t, filepath.Join(paths.Tags, "beta.yaml"), beta)
	collection := "name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - tag: alpha\n    - tag: beta\n"
	writeTestFile(t, filepath.Join(paths.Collections, "main.yaml"), collection)
	catalog, catalogDiagnostics := LoadCatalog(paths, DefaultAppConfig())
	if catalogDiagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", catalogDiagnostics)
	}
	graph, diagnostics := BuildValidatedGraph(catalog)
	if diagnostics.HasErrors() {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
	if len(graph.Demands(SourceCondition{Tag: "alpha"})) != 1 ||
		len(graph.Demands(SourceCondition{Tag: "beta"})) != 1 {
		t.Fatal("demand cycle was not preserved")
	}
	if len(graph.Suggestions(SourceCondition{Tag: "alpha"})) != 1 ||
		len(graph.Suggestions(SourceCondition{Tag: "beta"})) != 1 {
		t.Fatal("suggest cycle was not preserved")
	}
}
