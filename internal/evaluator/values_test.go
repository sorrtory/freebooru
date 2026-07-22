package evaluator

import (
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestAllowedValuesReportsOnlyNewViolations(t *testing.T) {
	catalog := testCatalog(t)
	graph, diagnostics := config.BuildValidatedGraph(catalog)
	if diagnostics.HasErrors() {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
	evaluator, err := New(catalog, graph)
	if err != nil {
		t.Fatal(err)
	}
	state, err := NewFileState(catalog, map[string]any{"reviewed": true})
	if err != nil {
		t.Fatal(err)
	}

	values, err := evaluator.AllowedValues("RATING", state)
	if err != nil {
		t.Fatalf("AllowedValues() error = %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("AllowedValues() = %#v", values)
	}
	if values[0].Value.Val != "safe" || !values[0].Allowed {
		t.Fatalf("safe availability = %#v", values[0])
	}
	if values[1].Value.Val != "explicit" || values[1].Allowed {
		t.Fatalf("explicit availability = %#v", values[1])
	}
	if len(values[1].Reasons) != 1 ||
		!strings.Contains(values[1].Reasons[0].Message, "conflicts with") {
		t.Fatalf("explicit reasons = %#v", values[1].Reasons)
	}
	if values[1].Reasons[0].Edge.Location.Field != "conflict[0]" {
		t.Fatalf("reason edge = %#v", values[1].Reasons[0].Edge)
	}
	if _, present := state.Value("rating"); present {
		t.Fatal("AllowedValues() mutated the original state")
	}
}

func TestAllowedValuesRejectsUnknownAndScalarTags(t *testing.T) {
	catalog := testCatalog(t)
	graph, diagnostics := config.BuildValidatedGraph(catalog)
	if diagnostics.HasErrors() {
		t.Fatalf("BuildValidatedGraph() diagnostics = %#v", diagnostics)
	}
	evaluator, err := New(catalog, graph)
	if err != nil {
		t.Fatal(err)
	}
	state, err := NewFileState(catalog, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := evaluator.AllowedValues("missing", state); err == nil {
		t.Fatal("AllowedValues(missing) error = nil")
	}
	if _, err := evaluator.AllowedValues("reviewed", state); err == nil {
		t.Fatal("AllowedValues(reviewed) error = nil")
	}
}
