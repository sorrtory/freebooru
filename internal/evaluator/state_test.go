package evaluator

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestNewFileStateValidatesAndCopiesAssignments(t *testing.T) {
	catalog := testCatalog(t)
	values := []string{"safe"}
	state, err := NewFileState(catalog, map[string]any{
		"REVIEWED": true,
		"rating":   values,
	})
	if err != nil {
		t.Fatalf("NewFileState() error = %v", err)
	}
	values[0] = "changed"
	got, ok := state.Value("RATING")
	if !ok || got.([]string)[0] != "safe" {
		t.Fatalf("Value(RATING) = %#v, %t", got, ok)
	}
	got.([]string)[0] = "changed again"
	again, _ := state.Value("rating")
	if again.([]string)[0] != "safe" {
		t.Fatalf("Value() exposed state: %#v", again)
	}
}

func TestNewFileStateTreatsBooleanFalseAsAbsent(t *testing.T) {
	state, err := NewFileState(testCatalog(t), map[string]any{"reviewed": false})
	if err != nil {
		t.Fatalf("NewFileState() error = %v", err)
	}
	if _, ok := state.Value("reviewed"); ok {
		t.Fatal("Value(reviewed) is present for false bool")
	}
	if sources := state.ActiveSources(); len(sources) != 0 {
		t.Fatalf("ActiveSources() = %#v", sources)
	}
}

func TestFileStateReturnsDeterministicActiveSources(t *testing.T) {
	state, err := NewFileState(testCatalog(t), map[string]any{
		"rating":   []string{"explicit", "safe"},
		"reviewed": true,
	})
	if err != nil {
		t.Fatalf("NewFileState() error = %v", err)
	}
	want := []config.SourceCondition{
		{Tag: "rating"},
		{Tag: "rating", Value: "explicit"},
		{Tag: "rating", Value: "safe"},
		{Tag: "reviewed"},
	}
	if got := state.ActiveSources(); !reflect.DeepEqual(got, want) {
		t.Fatalf("ActiveSources() = %#v, want %#v", got, want)
	}
}

func TestNewFileStateRejectsUnknownInvalidAndDuplicateTags(t *testing.T) {
	catalog := testCatalog(t)
	tests := []struct {
		name   string
		values map[string]any
		want   string
	}{
		{name: "unknown", values: map[string]any{"missing": true}, want: "does not exist"},
		{name: "invalid", values: map[string]any{"reviewed": "true"}, want: "expects"},
		{
			name:   "duplicate",
			values: map[string]any{"rating": []string{"safe"}, "RATING": []string{"safe"}},
			want:   "more than once",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFileState(catalog, tt.values)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("NewFileState() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func testCatalog(t *testing.T) *config.Catalog {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := config.PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	app, err := config.EnsureDefaults(paths)
	if err != nil {
		t.Fatal(err)
	}
	tags := "name: reviewed\ntype: bool\n---\n" +
		"name: rating\ntype: multivalue\nvalues:\n  - val: safe\n  - val: explicit\n"
	if err := os.WriteFile(filepath.Join(paths.Tags, "state.yaml"), []byte(tags), 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, diagnostics := config.LoadCatalog(paths, app)
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	return catalog
}
