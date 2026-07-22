package collection

import (
	"reflect"
	"strings"
	"testing"
)

func TestDatabaseSearchMatchesTypedANDTerms(t *testing.T) {
	database := openInitializedTestDatabase(t)
	seedSearchFiles(t, database)
	tests := []struct {
		name  string
		terms []SearchTerm
		want  []string
	}{
		{
			name:  "boolean presence",
			terms: []SearchTerm{{Tag: "reviewed", Type: "bool", Operator: SearchPresent}},
			want:  []string{"a", "c"},
		},
		{
			name:  "absence",
			terms: []SearchTerm{{Tag: "reviewed", Type: "bool", Operator: SearchAbsent}},
			want:  []string{"b"},
		},
		{
			name:  "boolean false",
			terms: []SearchTerm{{Tag: "reviewed", Type: "bool", Operator: SearchEqual, Value: false}},
			want:  []string{"b"},
		},
		{
			name:  "text equality",
			terms: []SearchTerm{{Tag: "title", Type: "text", Operator: SearchEqual, Value: "alpha"}},
			want:  []string{"a", "c"},
		},
		{
			name:  "integer range",
			terms: []SearchTerm{{Tag: "score", Type: "int", Operator: SearchGreaterEqual, Value: int64(20)}},
			want:  []string{"b", "c"},
		},
		{
			name:  "date ordering",
			terms: []SearchTerm{{Tag: "day", Type: "date", Operator: SearchGreater, Value: "2026-07-21"}},
			want:  []string{"b", "c"},
		},
		{
			name: "datetime instant equality",
			terms: []SearchTerm{{
				Tag: "instant", Type: "datetime", Operator: SearchEqual, Value: "2026-07-22T10:00:00Z",
			}},
			want: []string{"a", "b"},
		},
		{
			name: "datetime ordering",
			terms: []SearchTerm{{
				Tag: "instant", Type: "datetime", Operator: SearchGreater, Value: "2026-07-22T10:30:00Z",
			}},
			want: []string{"c"},
		},
		{
			name:  "predefined equality",
			terms: []SearchTerm{{Tag: "rating", Type: "value", Operator: SearchEqual, Value: "safe"}},
			want:  []string{"a", "c"},
		},
		{
			name:  "multivalue membership",
			terms: []SearchTerm{{Tag: "labels", Type: "multivalue", Operator: SearchEqual, Value: "second"}},
			want:  []string{"b", "c"},
		},
		{
			name:  "storage membership",
			terms: []SearchTerm{{Tag: "storage", Type: "multivalue", Operator: SearchEqual, Value: "default"}},
			want:  []string{"a", "c"},
		},
		{
			name: "AND combination",
			terms: []SearchTerm{
				{Tag: "reviewed", Type: "bool", Operator: SearchPresent},
				{Tag: "score", Type: "int", Operator: SearchGreaterEqual, Value: int64(20)},
			},
			want: []string{"c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := database.Search(t.Context(), SearchRequest{Terms: tt.terms, Limit: 100})
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}
			got := shortHashes(files)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Search() hashes = %#v, want %#v", got, tt.want)
			}
			for _, file := range files {
				if len(file.Tags) == 0 || len(file.Storages) == 0 || len(file.Sources) == 0 {
					t.Fatalf("Search() returned incomplete record %#v", file)
				}
			}
		})
	}
}

func TestDatabaseSearchOrdersAndPaginatesDeterministically(t *testing.T) {
	database := openInitializedTestDatabase(t)
	seedSearchFiles(t, database)
	files, err := database.Search(t.Context(), SearchRequest{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if got := shortHashes(files); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("Search() hashes = %#v, want [b]", got)
	}
	files, err = database.Search(t.Context(), SearchRequest{Limit: 0})
	if err != nil {
		t.Fatalf("Search(limit 0) error = %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("Search(limit 0) = %#v", files)
	}
}

func seedSearchFiles(t *testing.T, database *Database) {
	t.Helper()
	createSearchFile(t, database, "a", "alpha", 10, "2026-07-21", "2026-07-22T10:00:00Z", "safe", "first", true, "default")
	createSearchFile(t, database, "b", "beta", 20, "2026-07-22", "2026-07-22T12:00:00+02:00", "explicit", "second", false, "archive")
	createSearchFile(t, database, "c", "alpha", 30, "2026-07-23", "2026-07-22T11:00:00Z", "safe", "second", true, "default")
	for hash, importedAt := range map[string]string{
		strings.Repeat("a", 64): "2026-01-03 00:00:00",
		strings.Repeat("b", 64): "2026-01-02 00:00:00",
		strings.Repeat("c", 64): "2026-01-01 00:00:00",
	} {
		if _, err := database.db.ExecContext(
			t.Context(),
			"UPDATE file SET imported_at = ? WHERE sha256 = ?",
			importedAt,
			hash,
		); err != nil {
			t.Fatalf("set imported_at: %v", err)
		}
	}
}

func createSearchFile(
	t *testing.T,
	database *Database,
	hashCharacter string,
	title string,
	score int64,
	day string,
	instant string,
	rating string,
	label string,
	reviewed bool,
	storage string,
) {
	t.Helper()
	tags := []TagRecord{
		{Name: "title", Type: "text", TextValue: &title},
		{Name: "score", Type: "int", IntegerValue: &score},
		{Name: "day", Type: "date", TextValue: &day},
		{Name: "instant", Type: "datetime", TextValue: &instant},
		{Name: "rating", Type: "value", TextValue: &rating},
		{Name: "labels", Type: "multivalue", Values: []string{label}},
	}
	if reviewed {
		tags = append(tags, TagRecord{Name: "reviewed", Type: "bool"})
	}
	_, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         strings.Repeat(hashCharacter, 64),
		SizeBytes:      score,
		SourcePath:     "/imports/" + hashCharacter,
		SourceFilename: hashCharacter,
		Tags:           tags,
		Storages:       []string{storage},
	})
	if err != nil {
		t.Fatalf("CreateFile(%s) error = %v", hashCharacter, err)
	}
}

func shortHashes(files []FileRecord) []string {
	hashes := make([]string, 0, len(files))
	for _, file := range files {
		hashes = append(hashes, file.SHA256[:1])
	}
	return hashes
}
