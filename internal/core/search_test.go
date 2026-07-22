package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
)

func TestSearchResolvesDefaultCollectionAndTypedTerms(t *testing.T) {
	database := &fakeDatabase{searchFiles: []collection.FileRecord{{
		SHA256: strings.Repeat("a", 64),
	}}}
	app := newImportTestCoreWithDatabase(t, database)

	files, err := app.Search(t.Context(), FileSearchRequest{
		Terms:  []string{"RATING:SAFE", "score>=7"},
		Offset: 2,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(files) != 1 || database.searchRequest == nil {
		t.Fatalf("Search() files = %#v, request = %#v", files, database.searchRequest)
	}
	request := database.searchRequest
	if request.Limit != DefaultSearchLimit || request.Offset != 2 || len(request.Terms) != 2 {
		t.Fatalf("repository request = %#v", request)
	}
	if request.Terms[0].Tag != "rating" || request.Terms[0].Value != "safe" ||
		request.Terms[1].Value != int64(7) {
		t.Fatalf("repository terms = %#v", request.Terms)
	}
	if !database.closed {
		t.Fatal("Search() did not close database")
	}
}

func TestSearchPreservesExplicitZeroLimit(t *testing.T) {
	database := &fakeDatabase{}
	app := newImportTestCoreWithDatabase(t, database)
	limit := int64(0)

	if _, err := app.Search(t.Context(), FileSearchRequest{
		Terms: []string{"reviewed"}, Limit: &limit,
	}); err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if database.searchRequest == nil || database.searchRequest.Limit != 0 {
		t.Fatalf("repository request = %#v", database.searchRequest)
	}
}

func TestSearchRejectsInvalidInputBeforeOpeningDatabase(t *testing.T) {
	database := &fakeDatabase{}
	app := newImportTestCoreWithDatabase(t, database)
	tests := []struct {
		name    string
		request FileSearchRequest
	}{
		{name: "unsupported syntax", request: FileSearchRequest{Terms: []string{"a OR b"}}},
		{name: "unavailable tag", request: FileSearchRequest{Terms: []string{"missing"}}},
		{name: "negative offset", request: FileSearchRequest{Terms: []string{"reviewed"}, Offset: -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database.initialized = false
			if _, err := app.Search(t.Context(), tt.request); err == nil {
				t.Fatalf("Search(%#v) error = nil", tt.request)
			}
			if database.initialized {
				t.Fatal("Search() opened database for invalid input")
			}
		})
	}
}

func TestSearchReturnsRepositoryAndCloseErrors(t *testing.T) {
	database := &fakeDatabase{
		searchErr: errors.New("query failure"),
		closeErr:  errors.New("close failure"),
	}
	app := newImportTestCoreWithDatabase(t, database)

	_, err := app.Search(t.Context(), FileSearchRequest{Terms: []string{"reviewed"}})
	if err == nil || !strings.Contains(err.Error(), "query failure") ||
		!strings.Contains(err.Error(), "close failure") {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestSearchUsesResolvedTermsAgainstRealCollection(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	tags := `name: reviewed
type: bool
---
name: score
type: int
---
name: title
type: text
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
    - tag: score
    - tag: title
`
	if err := os.WriteFile(
		filepath.Join(paths.Collections, "main.yaml"),
		[]byte(collectionDocument),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	app, err := New(testLogger(), paths, func(ctx context.Context, path string) (CollectionDatabase, error) {
		return collection.Open(ctx, path)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	database, err := app.openCollectionDatabase(t.Context(), "main")
	if err != nil {
		t.Fatal(err)
	}
	for index, reviewed := range []bool{false, true} {
		hash := strings.Repeat(string(rune('a'+index)), 64)
		score := int64(10 + index*10)
		title := []string{"first", "second"}[index]
		tags := []collection.TagRecord{
			{Name: "score", Type: "int", IntegerValue: &score},
			{Name: "title", Type: "text", TextValue: &title},
		}
		if reviewed {
			tags = append(tags, collection.TagRecord{Name: "reviewed", Type: "bool"})
		}
		if _, err := database.CreateFile(t.Context(), collection.NewFile{
			SHA256:         hash,
			SizeBytes:      score,
			SourcePath:     "/imports/" + title,
			SourceFilename: title,
			Tags:           tags,
			Storages:       []string{"default"},
		}); err != nil {
			t.Fatalf("CreateFile(%s) error = %v", title, err)
		}
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	files, err := app.Search(t.Context(), FileSearchRequest{
		Terms: []string{"reviewed", "score>=20", "title:second"},
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(files) != 1 || files[0].SHA256 != strings.Repeat("b", 64) ||
		len(files[0].Sources) != 1 || files[0].Sources[0].Filename != "second" {
		t.Fatalf("Search() files = %#v", files)
	}
}
