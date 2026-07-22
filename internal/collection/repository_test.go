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
	})
	if err != nil {
		t.Fatalf("first CreateFile() error = %v", err)
	}
	_, err = database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      42,
		SourcePath:     "/imports/second.png",
		SourceFilename: "second.png",
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
	file, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      7,
		SourcePath:     "/imports/tagged.txt",
		SourceFilename: "tagged.txt",
	})
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	if _, err := database.db.ExecContext(
		t.Context(),
		`INSERT INTO file_tag (file_id, tag_name, tag_type, integer_value)
		 VALUES (?, ?, ?, ?)`,
		file.ID,
		"score",
		"int",
		int64(12),
	); err != nil {
		t.Fatalf("insert integer tag: %v", err)
	}
	result, err := database.db.ExecContext(
		t.Context(),
		`INSERT INTO file_tag (file_id, tag_name, tag_type)
		 VALUES (?, ?, ?)`,
		file.ID,
		"labels",
		"multivalue",
	)
	if err != nil {
		t.Fatalf("insert multivalue tag: %v", err)
	}
	tagID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read multivalue tag id: %v", err)
	}
	for _, value := range []string{"first", "second"} {
		if _, err := database.db.ExecContext(
			t.Context(),
			"INSERT INTO file_tag_value (file_tag_id, value) VALUES (?, ?)",
			tagID,
			value,
		); err != nil {
			t.Fatalf("insert multivalue %q: %v", value, err)
		}
	}
	if _, err := database.db.ExecContext(
		t.Context(),
		"INSERT INTO file_storage (file_id, storage_name) VALUES (?, ?)",
		file.ID,
		"default",
	); err != nil {
		t.Fatalf("insert storage: %v", err)
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

func openInitializedTestDatabase(t *testing.T) *Database {
	t.Helper()
	database := openTestDatabase(t)
	if err := database.Initialize(t.Context()); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	return database
}
