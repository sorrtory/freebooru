package collection

import (
	"errors"
	"strings"
	"testing"
)

func TestDatabaseCreateAndLoadFile(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("a", 64)
	created, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      42,
		SourcePath:     "/imports/example.png",
		SourceFilename: "example.png",
		Storages:       []string{"default"},
	})
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	if created.SHA256 != hash || created.SizeBytes != 42 {
		t.Fatalf("CreateFile() = %#v", created)
	}
	if len(created.Sources) != 1 || created.Sources[0].Path != "/imports/example.png" {
		t.Fatalf("CreateFile() sources = %#v", created.Sources)
	}
	loaded, err := database.File(t.Context(), hash)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	if loaded.ID != created.ID || loaded.SHA256 != created.SHA256 {
		t.Fatalf("File() = %#v, want identity %#v", loaded, created)
	}
}

func TestDatabaseCreateFileRollsBackSourceFailure(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("b", 64)
	_, err := database.CreateFile(t.Context(), NewFile{
		SHA256:    hash,
		SizeBytes: 42,
		Storages:  []string{"default"},
	})
	if err == nil {
		t.Fatal("CreateFile() error = nil, want source constraint error")
	}
	if _, err := database.File(t.Context(), hash); !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("File() error = %v, want ErrFileNotFound", err)
	}
}

func TestDatabaseCreateFileReturnsDuplicate(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("c", 64)
	first, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      42,
		SourcePath:     "/imports/first.png",
		SourceFilename: "first.png",
		Storages:       []string{"default"},
	})
	if err != nil {
		t.Fatalf("first CreateFile() error = %v", err)
	}
	_, err = database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      42,
		SourcePath:     "/imports/second.png",
		SourceFilename: "second.png",
		Storages:       []string{"archive"},
	})
	if !errors.Is(err, ErrDuplicateFile) {
		t.Fatalf("second CreateFile() error = %v, want ErrDuplicateFile", err)
	}
	var duplicate *DuplicateFileError
	if !errors.As(err, &duplicate) {
		t.Fatalf("second CreateFile() error type = %T, want *DuplicateFileError", err)
	}
	if duplicate.Existing.ID != first.ID || len(duplicate.Existing.Sources) != 1 {
		t.Fatalf("duplicate existing file = %#v", duplicate.Existing)
	}
}

func TestDatabaseFileReturnsNotFound(t *testing.T) {
	database := openInitializedTestDatabase(t)
	_, err := database.File(t.Context(), strings.Repeat("d", 64))
	if !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("File() error = %v, want ErrFileNotFound", err)
	}
}

func TestDatabaseFileLoadsTypedTagsAndStorages(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("e", 64)
	score := int64(12)
	_, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      7,
		SourcePath:     "/imports/tagged.txt",
		SourceFilename: "tagged.txt",
		Tags: []TagRecord{
			{Name: "labels", Type: "multivalue", Values: []string{"first", "second"}},
			{Name: "score", Type: "int", IntegerValue: &score},
		},
		Storages: []string{"default"},
	})
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	loaded, err := database.File(t.Context(), hash)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	if len(loaded.Tags) != 2 {
		t.Fatalf("File() tags = %#v, want 2 tags", loaded.Tags)
	}
	if loaded.Tags[0].Name != "labels" || len(loaded.Tags[0].Values) != 2 {
		t.Errorf("File() multivalue tag = %#v", loaded.Tags[0])
	}
	if loaded.Tags[1].Name != "score" || loaded.Tags[1].IntegerValue == nil ||
		*loaded.Tags[1].IntegerValue != 12 {
		t.Errorf("File() integer tag = %#v", loaded.Tags[1])
	}
	if len(loaded.Storages) != 1 || loaded.Storages[0] != "default" {
		t.Errorf("File() storages = %#v", loaded.Storages)
	}
}

func TestDatabaseStorageMutations(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("f", 64)
	_, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      9,
		SourcePath:     "/imports/stored.txt",
		SourceFilename: "stored.txt",
		Tags:           []TagRecord{{Name: "reviewed", Type: "bool"}},
		Storages:       []string{"default"},
	})
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	change, err := database.AddStorage(t.Context(), hash, "archive")
	if err != nil {
		t.Fatalf("AddStorage() error = %v", err)
	}
	if !change.Changed || change.FileDeleted {
		t.Fatalf("AddStorage() = %#v", change)
	}
	change, err = database.AddStorage(t.Context(), hash, "archive")
	if err != nil {
		t.Fatalf("repeated AddStorage() error = %v", err)
	}
	if change.Changed {
		t.Fatalf("repeated AddStorage() = %#v, want no-op", change)
	}
	change, err = database.RemoveStorage(t.Context(), hash, "missing")
	if err != nil {
		t.Fatalf("missing RemoveStorage() error = %v", err)
	}
	if change.Changed {
		t.Fatalf("missing RemoveStorage() = %#v, want no-op", change)
	}
	change, err = database.RemoveStorage(t.Context(), hash, "archive")
	if err != nil {
		t.Fatalf("RemoveStorage() error = %v", err)
	}
	if !change.Changed || change.FileDeleted {
		t.Fatalf("RemoveStorage() = %#v", change)
	}
	change, err = database.RemoveStorage(t.Context(), hash, "default")
	if err != nil {
		t.Fatalf("final RemoveStorage() error = %v", err)
	}
	if !change.Changed || !change.FileDeleted {
		t.Fatalf("final RemoveStorage() = %#v", change)
	}
	if _, err := database.File(t.Context(), hash); !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("File() after final removal error = %v, want ErrFileNotFound", err)
	}
	for _, table := range []string{"file_source", "file_tag", "file_storage"} {
		var count int
		if err := database.db.QueryRowContext(
			t.Context(),
			"SELECT COUNT(*) FROM "+table, //nolint:gosec // fixed test table allowlist.
		).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Errorf("%s row count = %d, want 0", table, count)
		}
	}
}

func TestDatabaseStorageMutationsReturnNotFound(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("1", 64)
	if _, err := database.AddStorage(t.Context(), hash, "default"); !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("AddStorage() error = %v, want ErrFileNotFound", err)
	}
	if _, err := database.RemoveStorage(t.Context(), hash, "default"); !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("RemoveStorage() error = %v, want ErrFileNotFound", err)
	}
}

func openInitializedTestDatabase(t *testing.T) *Database {
	t.Helper()
	database := openTestDatabase(t)
	if err := database.Initialize(t.Context()); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	return database
}
