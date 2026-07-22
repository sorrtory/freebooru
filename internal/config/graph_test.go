package config

import (
	"path/filepath"
	"testing"
)

func TestBuildGraphCreatesDeterministicDirectedIndexes(t *testing.T) {
	paths := writeCatalogFixture(t)
	artist := "name: artist\ntype: text\ngroups: [content]\n" +
		"suggest:\n  - tag: rating\n    has: [safe]\n    reason: useful\n" +
		"demand:\n  - tag: reviewed\n" +
		"conflict:\n  - tag: blocked\n"
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), artist)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}

	graph := BuildGraph(catalog)
	source := SourceCondition{Tag: "ARTIST"}
	edges := graph.Outgoing(source)
	if len(edges) != 3 {
		t.Fatalf("Outgoing(artist) = %#v", edges)
	}
	wantKinds := []RelationshipKind{
		RelationshipSuggest,
		RelationshipDemand,
		RelationshipConflict,
	}
	for index, want := range wantKinds {
		if edges[index].Kind != want {
			t.Fatalf("edge[%d].Kind = %q, want %q", index, edges[index].Kind, want)
		}
	}
	if edges[0].Reason != "useful" || edges[0].Location.Field != "suggest[0]" {
		t.Fatalf("suggestion edge = %#v", edges[0])
	}
	if got := graph.Backlinks("RATING"); len(got) != 1 || got[0].Source.Tag != "artist" {
		t.Fatalf("Backlinks(RATING) = %#v", got)
	}
	if got := graph.Outgoing(SourceCondition{Tag: "rating"}); len(got) != 0 {
		t.Fatalf("Outgoing(rating) inferred reverse edges: %#v", got)
	}
}

func TestBuildGraphIndexesValueLevelRelationships(t *testing.T) {
	paths := writeCatalogFixture(t)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	graph := BuildGraph(catalog)
	source := SourceCondition{Tag: "RATING", Value: "SAFE"}
	edges := graph.Demands(source)
	if len(edges) != 1 || edges[0].TargetTag != "reviewed" {
		t.Fatalf("Demands(rating:safe) = %#v", edges)
	}
	if edges[0].Location.Field != "values[0].demand[0]" {
		t.Fatalf("edge location = %#v", edges[0].Location)
	}
}

func TestGraphQueriesReturnDefensivePredicateCopies(t *testing.T) {
	paths := writeCatalogFixture(t)
	artist := "name: artist\ntype: text\nsuggest:\n  - tag: storage\n    has: [default]\n"
	writeTestFile(t, filepath.Join(paths.Tags, "artist.yaml"), artist)
	catalog, diagnostics := LoadCatalog(paths, DefaultAppConfig())
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	graph, compileDiagnostics := CompileGraph(catalog, BuildGraph(catalog))
	if compileDiagnostics.HasErrors() {
		t.Fatalf("CompileGraph() diagnostics = %#v", compileDiagnostics)
	}
	source := SourceCondition{Tag: "artist"}
	edges := graph.Suggestions(source)
	edges[0].Predicate.Has[0] = "changed"
	again := graph.Suggestions(source)
	if again[0].Predicate.Has[0] != "default" {
		t.Fatalf("Suggestions() exposed graph predicate: %#v", again)
	}
}
