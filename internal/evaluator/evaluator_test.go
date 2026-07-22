package evaluator

import (
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestEvaluatorReturnsMissingDemandsConflictsAndSuggestions(t *testing.T) {
	catalog := testCatalog(t)
	graph, diagnostics := config.BuildValidatedGraph(catalog)
	if diagnostics.HasErrors() {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
	evaluator, err := New(catalog, graph)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	state, err := NewFileState(catalog, map[string]any{
		"reviewed": true,
		"rating":   []string{"explicit"},
	})
	if err != nil {
		t.Fatalf("NewFileState() error = %v", err)
	}

	result := evaluator.ValidateFile(state)
	if result.Valid() {
		t.Fatal("ValidateFile() result is valid")
	}
	if len(result.MissingDemands) != 1 ||
		result.MissingDemands[0].TargetTag != "rating" {
		t.Fatalf("MissingDemands = %#v", result.MissingDemands)
	}
	if len(result.ActiveConflicts) != 1 ||
		result.ActiveConflicts[0].TargetTag != "rating" {
		t.Fatalf("ActiveConflicts = %#v", result.ActiveConflicts)
	}
	if len(result.Suggestions) != 1 ||
		result.Suggestions[0].TargetTag != "rating" {
		t.Fatalf("Suggestions = %#v", result.Suggestions)
	}
}

func TestEvaluatorOmitsSatisfiedSuggestionsAndDemands(t *testing.T) {
	catalog := testCatalog(t)
	graph, diagnostics := config.BuildValidatedGraph(catalog)
	if diagnostics.HasErrors() {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
	evaluator, err := New(catalog, graph)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	state, err := NewFileState(catalog, map[string]any{
		"reviewed": true,
		"rating":   []string{"safe"},
	})
	if err != nil {
		t.Fatalf("NewFileState() error = %v", err)
	}

	result := evaluator.ValidateFile(state)
	if !result.Valid() {
		t.Fatalf("ValidateFile() result = %#v", result)
	}
	if len(result.Suggestions) != 0 {
		t.Fatalf("Suggestions = %#v", result.Suggestions)
	}
}

func TestEvaluatorConvenienceQueriesMatchValidation(t *testing.T) {
	catalog := testCatalog(t)
	graph, diagnostics := config.BuildValidatedGraph(catalog)
	if diagnostics.HasErrors() {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
	evaluator, err := New(catalog, graph)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	state, err := NewFileState(catalog, map[string]any{
		"reviewed": true,
		"rating":   []string{"explicit"},
	})
	if err != nil {
		t.Fatalf("NewFileState() error = %v", err)
	}
	if len(evaluator.MissingDemands(state)) != 1 {
		t.Fatal("MissingDemands() did not return the active demand")
	}
	if len(evaluator.ActiveConflicts(state)) != 1 {
		t.Fatal("ActiveConflicts() did not return the active conflict")
	}
	if len(evaluator.Suggestions(state)) != 1 {
		t.Fatal("Suggestions() did not return the active suggestion")
	}
}

func TestNewEvaluatorRequiresDependencies(t *testing.T) {
	catalog := testCatalog(t)
	graph, diagnostics := config.BuildValidatedGraph(catalog)
	if diagnostics.HasErrors() {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
	if _, err := New(nil, graph); err == nil {
		t.Fatal("New(nil, graph) error = nil")
	}
	if _, err := New(catalog, nil); err == nil {
		t.Fatal("New(catalog, nil) error = nil")
	}
}
