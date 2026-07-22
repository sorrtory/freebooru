package search

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

func TestResolveCanonicalizesAndTypesEverySupportedOperand(t *testing.T) {
	catalog, references := searchTestCatalog(t)
	query, err := Parse([]string{
		"REVIEWED",
		"!title",
		"reviewed:false",
		"title:hello world",
		"score>=10",
		"day<2026-07-22",
		"instant>2026-07-22T10:30:00Z",
		"rating:SFW",
		"labels:TWO",
		"storage:DEFAULT",
		"filesize>=512",
		"filetype:image/png",
		"imported_at>2026-07-22T10:30:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := Resolve(query, catalog, references)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := []ResolvedTerm{
		{Tag: "reviewed", Type: config.TagTypeBool, Operator: Present},
		{Tag: "title", Type: config.TagTypeText, Operator: Absent},
		{Tag: "reviewed", Type: config.TagTypeBool, Operator: Equal, Value: false},
		{Tag: "title", Type: config.TagTypeText, Operator: Equal, Value: "hello world"},
		{Tag: "score", Type: config.TagTypeInt, Operator: GreaterEqual, Value: int64(10)},
		{Tag: "day", Type: config.TagTypeDate, Operator: Less, Value: "2026-07-22"},
		{
			Tag: "instant", Type: config.TagTypeDatetime, Operator: Greater,
			Value: "2026-07-22T10:30:00Z",
		},
		{Tag: "rating", Type: config.TagTypeValue, Operator: Equal, Value: "safe"},
		{Tag: "labels", Type: config.TagTypeMultivalue, Operator: Equal, Value: "second"},
		{Tag: "storage", Type: config.TagTypeMultivalue, Operator: Equal, Value: "default"},
		{Tag: "filesize", Type: config.TagTypeInt, Operator: GreaterEqual, Value: int64(512)},
		{Tag: "filetype", Type: config.TagTypeText, Operator: Equal, Value: "image/png"},
		{
			Tag: "imported_at", Type: config.TagTypeDatetime, Operator: Greater,
			Value: "2026-07-22T10:30:00Z",
		},
	}
	if !reflect.DeepEqual(resolved.Terms, want) {
		t.Fatalf("Resolve() terms = %#v, want %#v", resolved.Terms, want)
	}
}

func TestResolveRejectsUnavailableAndTypeIncompatibleTerms(t *testing.T) {
	catalog, references := searchTestCatalog(t)
	tests := []struct {
		name string
		term string
		want string
	}{
		{name: "unimported tag", term: "hidden", want: "not imported"},
		{name: "unimported storage", term: "storage:archive", want: "not imported"},
		{name: "presence on text", term: "title", want: "boolean"},
		{name: "order on text", term: "title>hello", want: "ordered"},
		{name: "invalid integer", term: "score>=ten", want: "integer"},
		{name: "negative integer", term: "score>=-1", want: "at least 0"},
		{name: "invalid date", term: "day<tomorrow", want: "2006-01-02"},
		{name: "unknown value", term: "rating:unknown", want: "does not declare"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := Parse([]string{tt.term})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Resolve(query, catalog, references)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Resolve(%q) error = %v, want %q", tt.term, err, tt.want)
			}
		})
	}
}

func searchTestCatalog(t *testing.T) (*config.Catalog, config.ResolvedReferences) {
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
	tags := `name: reviewed
type: bool
---
name: title
type: text
---
name: score
type: int
---
name: day
type: date
---
name: instant
type: datetime
---
name: rating
type: value
values:
  - val: safe
    aliases: [sfw]
---
name: labels
type: multivalue
values:
  - val: first
  - val: second
    aliases: [two]
---
name: hidden
type: bool
`
	if err := os.WriteFile(filepath.Join(paths.Tags, "search.yaml"), []byte(tags), 0o600); err != nil {
		t.Fatal(err)
	}
	collectionDocument := `name: main
tags:
  require:
    - storage: default
  import:
    - tag: reviewed
    - tag: title
    - tag: score
    - tag: day
    - tag: instant
    - tag: rating
    - tag: labels
`
	if err := os.WriteFile(
		filepath.Join(paths.Collections, "main.yaml"),
		[]byte(collectionDocument),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	catalog, diagnostics := config.LoadCatalog(paths, app)
	if diagnostics.HasErrors() {
		t.Fatalf("LoadCatalog() diagnostics = %#v", diagnostics)
	}
	references, ok := catalog.CollectionReferences("main")
	if !ok {
		t.Fatal("main collection references are unavailable")
	}
	return catalog, references
}
