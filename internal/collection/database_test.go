package collection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenRejectsRelativePath(t *testing.T) {
	_, err := Open(t.Context(), "collection.sqlite")
	if err == nil || !strings.Contains(err.Error(), "relative") {
		t.Fatalf("Open() error = %v, want relative path error", err)
	}
}

func TestDatabaseLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "main.sqlite")
	database, err := Open(t.Context(), path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := database.Initialize(t.Context()); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if err := database.Initialize(t.Context()); err != nil {
		t.Fatalf("second Initialize() error = %v", err)
	}
	if err := database.Ping(t.Context()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read SQLite file: %v", err)
	}
	if len(contents) < 16 || string(contents[:16]) != "SQLite format 3\x00" {
		t.Fatalf("database header = %q, want SQLite format 3", contents[:min(len(contents), 16)])
	}

	database, err = Open(t.Context(), path)
	if err != nil {
		t.Fatalf("reopen: Open() error = %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()
	version, err := schemaVersion(t.Context(), database.db)
	if err != nil {
		t.Fatalf("schemaVersion() error = %v", err)
	}
	if version != currentSchemaVersion {
		t.Fatalf("schemaVersion() = %d, want %d", version, currentSchemaVersion)
	}
	var foreignKeys bool
	if err := database.db.QueryRowContext(t.Context(), "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("read foreign_keys: %v", err)
	}
	if !foreignKeys {
		t.Fatal("foreign keys are disabled")
	}
}

func TestDatabaseRejectsNewerSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.sqlite")
	database, err := Open(t.Context(), path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()
	if err := database.Initialize(t.Context()); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if _, err := database.db.ExecContext(
		t.Context(),
		"INSERT INTO schema_migrations (version) VALUES (?)",
		currentSchemaVersion+1,
	); err != nil {
		t.Fatalf("insert newer schema version: %v", err)
	}
	err = database.Initialize(t.Context())
	if err == nil || !strings.Contains(err.Error(), "newer than supported") {
		t.Fatalf("Initialize() error = %v, want newer schema error", err)
	}
}
