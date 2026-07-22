package collection

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestDatabaseForEachFile(t *testing.T) {
	database := openInitializedTestDatabase(t)
	for _, character := range []string{"c", "a", "b"} {
		hash := strings.Repeat(character, 64)
		if _, err := database.CreateFile(t.Context(), NewFile{
			SHA256:         hash,
			SizeBytes:      1,
			SourcePath:     "/imports/" + character,
			SourceFilename: character,
			Storages:       []string{"default"},
		}); err != nil {
			t.Fatalf("CreateFile(%q) error = %v", character, err)
		}
	}
	var hashes []string
	if err := database.ForEachFile(t.Context(), func(file FileRecord) error {
		hashes = append(hashes, file.SHA256)
		if len(file.Sources) != 1 || len(file.Storages) != 1 {
			t.Errorf("visited file = %#v", file)
		}
		return nil
	}); err != nil {
		t.Fatalf("ForEachFile() error = %v", err)
	}
	want := []string{
		strings.Repeat("a", 64),
		strings.Repeat("b", 64),
		strings.Repeat("c", 64),
	}
	if !reflect.DeepEqual(hashes, want) {
		t.Fatalf("ForEachFile() hashes = %q, want %q", hashes, want)
	}
}

func TestDatabaseForEachFileStopsAtVisitorError(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("a", 64)
	if _, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      1,
		SourcePath:     "/imports/a",
		SourceFilename: "a",
		Storages:       []string{"default"},
	}); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	wantErr := errors.New("invalid persisted state")
	err := database.ForEachFile(t.Context(), func(FileRecord) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("ForEachFile() error = %v, want visitor error", err)
	}
}

func TestDatabaseForEachFileRejectsNilVisitor(t *testing.T) {
	database := openInitializedTestDatabase(t)
	if err := database.ForEachFile(t.Context(), nil); err == nil {
		t.Fatal("ForEachFile() error = nil, want nil callback error")
	}
}
