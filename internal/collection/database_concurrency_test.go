package collection

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIndependentDatabaseConnectionsShareCommittedState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.sqlite")
	gui := openDatabaseAt(t, path)
	if err := gui.Initialize(t.Context()); err != nil {
		t.Fatalf("GUI Initialize() error = %v", err)
	}
	cli := openDatabaseAt(t, path)
	if err := cli.Initialize(t.Context()); err != nil {
		t.Fatalf("CLI Initialize() error = %v", err)
	}
	hash := strings.Repeat("a", 64)
	if _, err := gui.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      1,
		SourcePath:     "/imports/shared",
		SourceFilename: "shared",
		Storages:       []string{"default"},
	}); err != nil {
		t.Fatalf("GUI CreateFile() error = %v", err)
	}
	if _, err := cli.File(t.Context(), hash); err != nil {
		t.Fatalf("CLI File() error = %v", err)
	}
	if _, err := cli.AddStorage(t.Context(), hash, "archive"); err != nil {
		t.Fatalf("CLI AddStorage() error = %v", err)
	}
	file, err := gui.File(t.Context(), hash)
	if err != nil {
		t.Fatalf("GUI File() after CLI write error = %v", err)
	}
	if len(file.Storages) != 2 {
		t.Fatalf("GUI File() storages = %q, want shared CLI commit", file.Storages)
	}
}

func TestIndependentWriterHonorsContextWhileDatabaseIsBusy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.sqlite")
	gui := openDatabaseAt(t, path)
	if err := gui.Initialize(t.Context()); err != nil {
		t.Fatalf("GUI Initialize() error = %v", err)
	}
	cli := openDatabaseAt(t, path)
	hash := strings.Repeat("b", 64)
	if _, err := gui.CreateFile(t.Context(), NewFile{
		SHA256:         hash,
		SizeBytes:      1,
		SourcePath:     "/imports/busy",
		SourceFilename: "busy",
		Storages:       []string{"default"},
	}); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}

	lock, err := gui.db.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatalf("begin GUI transaction: %v", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	_, err = cli.AddStorage(ctx, hash, "archive")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("contended AddStorage() error = %v, want context deadline", err)
	}
	if err := lock.Rollback(); err != nil {
		t.Fatalf("rollback GUI transaction: %v", err)
	}
	if _, err := cli.AddStorage(t.Context(), hash, "archive"); err != nil {
		t.Fatalf("AddStorage() after lock release error = %v", err)
	}
	file, err := gui.File(t.Context(), hash)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	if len(file.Storages) != 2 {
		t.Fatalf("File() storages = %q, want retry commit", file.Storages)
	}
}

func openDatabaseAt(t *testing.T, path string) *Database {
	t.Helper()
	database, err := Open(t.Context(), path)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", path, err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close(%q) error = %v", path, err)
		}
	})
	return database
}
