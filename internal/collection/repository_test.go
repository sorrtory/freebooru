package collection

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
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

func TestExplicitMutationTouchesInteractionButReadDoesNot(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("9", 64)
	if _, err := database.CreateFile(t.Context(), NewFile{
		SHA256: hash, SizeBytes: 1, SourcePath: "/imports/touch",
		SourceFilename: "touch", Storages: []string{"default"},
	}); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	const old = "2000-01-01 00:00:00"
	if _, err := database.db.ExecContext(
		t.Context(),
		"UPDATE file SET updated_at = ?, last_interaction_at = ? WHERE sha256 = ?",
		old,
		old,
		hash,
	); err != nil {
		t.Fatalf("set old interaction time: %v", err)
	}
	if _, err := database.File(t.Context(), hash); err != nil {
		t.Fatalf("passive File() error = %v", err)
	}
	passive, err := database.File(t.Context(), hash)
	if err != nil {
		t.Fatal(err)
	}
	if passive.LastInteractionAt != old {
		t.Fatalf("passive read changed interaction time to %q", passive.LastInteractionAt)
	}
	if _, err := database.AddTag(t.Context(), hash, TagRecord{Name: "reviewed", Type: "bool"}); err != nil {
		t.Fatalf("AddTag() error = %v", err)
	}
	mutated, err := database.File(t.Context(), hash)
	if err != nil {
		t.Fatal(err)
	}
	if mutated.LastInteractionAt == old || mutated.UpdatedAt == old {
		t.Fatalf("mutation timestamps = %#v", mutated)
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
	title := "example"
	day := "2026-07-22"
	instant := "2026-07-22T12:30:00Z"
	rating := "safe"
	_, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      7,
		SourcePath:     "/imports/tagged.txt",
		SourceFilename: "tagged.txt",
		Tags: []TagRecord{
			{Name: "published", Type: "bool"},
			{Name: "title", Type: "text", TextValue: &title},
			{Name: "labels", Type: "multivalue", Values: []string{"first", "second"}},
			{Name: "score", Type: "int", IntegerValue: &score},
			{Name: "day", Type: "date", TextValue: &day},
			{Name: "instant", Type: "datetime", TextValue: &instant},
			{Name: "rating", Type: "value", TextValue: &rating},
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
	if len(loaded.Tags) != 7 {
		t.Fatalf("File() tags = %#v, want 7 tags", loaded.Tags)
	}
	tags := tagsByName(loaded.Tags)
	if len(tags["labels"].Values) != 2 {
		t.Errorf("File() multivalue tag = %#v", tags["labels"])
	}
	if tags["score"].IntegerValue == nil || *tags["score"].IntegerValue != 12 {
		t.Errorf("File() integer tag = %#v", tags["score"])
	}
	for name, want := range map[string]string{
		"title": title, "day": day, "instant": instant, "rating": rating,
	} {
		if tags[name].TextValue == nil || *tags[name].TextValue != want {
			t.Errorf("File() %s tag = %#v, want %q", name, tags[name], want)
		}
	}
	if tags["published"].Type != "bool" {
		t.Errorf("File() boolean tag = %#v", tags["published"])
	}
	if len(loaded.Storages) != 1 || loaded.Storages[0] != "default" {
		t.Errorf("File() storages = %#v", loaded.Storages)
	}
}

func TestDatabaseSupportsConcurrentReadersAndSerializedWriters(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("4", 64)
	if _, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      4,
		SourcePath:     "/imports/concurrent.txt",
		SourceFilename: "concurrent.txt",
		Storages:       []string{"default"},
	}); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	start := make(chan struct{})
	errorsFound := make(chan error, 6)
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			<-start
			_, err := database.File(ctx, hash)
			errorsFound <- err
		})
	}
	for _, storage := range []string{"archive", "backup"} {
		group.Go(func() {
			<-start
			_, err := database.AddStorage(ctx, hash, storage)
			errorsFound <- err
		})
	}
	close(start)
	group.Wait()
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatalf("concurrent repository operation: %v", err)
		}
	}
	file, err := database.File(t.Context(), hash)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	if len(file.Storages) != 3 {
		t.Fatalf("File() storages = %q, want three serialized assignments", file.Storages)
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

func TestDatabaseTagMutations(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("2", 64)
	scoreOne := int64(1)
	_, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      11,
		SourcePath:     "/imports/tags.txt",
		SourceFilename: "tags.txt",
		Tags: []TagRecord{
			{Name: "labels", Type: "multivalue", Values: []string{"first"}},
			{Name: "score", Type: "int", IntegerValue: &scoreOne},
		},
		Storages: []string{"default"},
	})
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	change, err := database.AddTag(t.Context(), hash, TagRecord{
		Name: "score", Type: "int", IntegerValue: &scoreOne,
	})
	if err != nil {
		t.Fatalf("identical AddTag() error = %v", err)
	}
	if change.Changed {
		t.Fatalf("identical AddTag() = %#v, want no-op", change)
	}
	scoreTwo := int64(2)
	if _, err := database.AddTag(t.Context(), hash, TagRecord{
		Name: "score", Type: "int", IntegerValue: &scoreTwo,
	}); !errors.Is(err, ErrTagAlreadyAssigned) {
		t.Fatalf("replacement AddTag() error = %v, want ErrTagAlreadyAssigned", err)
	}
	change, err = database.AddTag(t.Context(), hash, TagRecord{
		Name: "labels", Type: "multivalue", Values: []string{"first", "second"},
	})
	if err != nil {
		t.Fatalf("multivalue AddTag() error = %v", err)
	}
	if !change.Changed {
		t.Fatalf("multivalue AddTag() = %#v, want change", change)
	}
	if _, err := database.SetTag(t.Context(), hash, TagRecord{
		Name: "score", Type: "int", IntegerValue: &scoreTwo,
	}); err != nil {
		t.Fatalf("SetTag() error = %v", err)
	}
	change, err = database.RemoveTag(t.Context(), hash, "labels")
	if err != nil {
		t.Fatalf("RemoveTag() error = %v", err)
	}
	if !change.Changed {
		t.Fatalf("RemoveTag() = %#v, want change", change)
	}
	change, err = database.RemoveTag(t.Context(), hash, "labels")
	if err != nil {
		t.Fatalf("repeated RemoveTag() error = %v", err)
	}
	if change.Changed {
		t.Fatalf("repeated RemoveTag() = %#v, want no-op", change)
	}
	loaded, err := database.File(t.Context(), hash)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	if len(loaded.Tags) != 1 || loaded.Tags[0].Name != "score" ||
		loaded.Tags[0].IntegerValue == nil || *loaded.Tags[0].IntegerValue != 2 {
		t.Fatalf("File() tags = %#v", loaded.Tags)
	}
}

func TestDatabaseTagMutationRejectsInvalidRecord(t *testing.T) {
	database := openInitializedTestDatabase(t)
	hash := strings.Repeat("3", 64)
	_, err := database.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      3,
		SourcePath:     "/imports/invalid.txt",
		SourceFilename: "invalid.txt",
		Storages:       []string{"default"},
	})
	if err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	text := "not an integer"
	if _, err := database.AddTag(t.Context(), hash, TagRecord{
		Name: "score", Type: "int", TextValue: &text,
	}); err == nil {
		t.Fatal("AddTag() error = nil, want invalid record error")
	}
	loaded, err := database.File(t.Context(), hash)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	if len(loaded.Tags) != 0 {
		t.Fatalf("File() tags = %#v, want no assignments", loaded.Tags)
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

func tagsByName(tags []TagRecord) map[string]TagRecord {
	indexed := make(map[string]TagRecord, len(tags))
	for _, tag := range tags {
		indexed[tag.Name] = tag
	}
	return indexed
}
