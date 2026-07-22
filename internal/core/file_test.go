package core

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
)

func TestGetFileUsesDefaultCollectionAndClosesDatabase(t *testing.T) {
	hash := strings.Repeat("a", 64)
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256:    hash,
		SizeBytes: 42,
		Storages:  []string{"default"},
	}}}
	app := newCheckedTestCore(t, database)

	record, err := app.GetFile(t.Context(), "", hash)
	if err != nil {
		t.Fatalf("GetFile() error = %v", err)
	}
	if record.SHA256 != hash || record.SizeBytes != 42 {
		t.Fatalf("GetFile() = %#v", record)
	}
	if !database.closed {
		t.Fatal("GetFile() did not close database")
	}
}

func TestGetFilePreservesNotFoundAndCloseErrors(t *testing.T) {
	database := &fakeDatabase{closeErr: errors.New("close failure")}
	app := newCheckedTestCore(t, database)

	_, err := app.GetFile(t.Context(), "main", strings.Repeat("b", 64))
	if !errors.Is(err, collection.ErrFileNotFound) {
		t.Fatalf("GetFile() error = %v, want ErrFileNotFound", err)
	}
	if !strings.Contains(err.Error(), "close failure") {
		t.Fatalf("GetFile() error = %v, want close failure", err)
	}
	if !database.closed {
		t.Fatal("GetFile() did not close database after lookup failure")
	}
}

func newCheckedTestCore(t *testing.T, database CollectionDatabase) *Core {
	t.Helper()
	paths, _ := provisionTestConfig(t)
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
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
	return app
}
