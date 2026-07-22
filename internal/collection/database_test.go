package collection

import (
	"database/sql"
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

func TestDatabaseUpgradesSchemaOne(t *testing.T) {
	database := openTestDatabase(t)
	if _, err := database.db.ExecContext(t.Context(), initialMigration); err != nil {
		t.Fatalf("apply migration 1: %v", err)
	}
	if err := database.Initialize(t.Context()); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	version, err := schemaVersion(t.Context(), database.db)
	if err != nil {
		t.Fatalf("schemaVersion() error = %v", err)
	}
	if version != currentSchemaVersion {
		t.Fatalf("schemaVersion() = %d, want %d", version, currentSchemaVersion)
	}
	for _, table := range []string{
		"file",
		"file_source",
		"file_tag",
		"file_tag_value",
		"file_storage",
		"file_relationship",
	} {
		if !tableExists(t, database.db, table) {
			t.Errorf("table %q does not exist", table)
		}
	}
}

func TestCollectionStateConstraintsAndCascades(t *testing.T) {
	database := openTestDatabase(t)
	if err := database.Initialize(t.Context()); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	validHash := strings.Repeat("a", 64)
	result, err := database.db.ExecContext(
		t.Context(),
		"INSERT INTO file (sha256, size_bytes) VALUES (?, ?)",
		validHash,
		12,
	)
	if err != nil {
		t.Fatalf("insert file: %v", err)
	}
	fileID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read inserted file id: %v", err)
	}
	assertStatementFails(
		t,
		database.db,
		"INSERT INTO file (sha256, size_bytes) VALUES (?, ?)",
		strings.Repeat("A", 64),
		12,
	)
	assertStatementFails(
		t,
		database.db,
		"INSERT INTO file (sha256, size_bytes) VALUES (?, ?)",
		strings.Repeat("b", 64),
		-1,
	)
	assertStatementFails(
		t,
		database.db,
		"INSERT INTO file_tag (file_id, tag_name, tag_type, text_value) VALUES (?, ?, ?, ?)",
		fileID,
		"score",
		"int",
		"12",
	)
	if _, err := database.db.ExecContext(
		t.Context(),
		"INSERT INTO file_source (file_id, source_path, filename) VALUES (?, ?, ?)",
		fileID,
		"/tmp/source.png",
		"source.png",
	); err != nil {
		t.Fatalf("insert file source: %v", err)
	}
	tagResult, err := database.db.ExecContext(
		t.Context(),
		"INSERT INTO file_tag (file_id, tag_name, tag_type) VALUES (?, ?, ?)",
		fileID,
		"labels",
		"multivalue",
	)
	if err != nil {
		t.Fatalf("insert file tag: %v", err)
	}
	tagID, err := tagResult.LastInsertId()
	if err != nil {
		t.Fatalf("read inserted tag id: %v", err)
	}
	if _, err := database.db.ExecContext(
		t.Context(),
		"INSERT INTO file_tag_value (file_tag_id, value) VALUES (?, ?)",
		tagID,
		"example",
	); err != nil {
		t.Fatalf("insert file tag value: %v", err)
	}
	if _, err := database.db.ExecContext(
		t.Context(),
		"INSERT INTO file_storage (file_id, storage_name) VALUES (?, ?)",
		fileID,
		"default",
	); err != nil {
		t.Fatalf("insert file storage: %v", err)
	}
	if _, err := database.db.ExecContext(
		t.Context(),
		"DELETE FROM file WHERE file_id = ?",
		fileID,
	); err != nil {
		t.Fatalf("delete file: %v", err)
	}
	for _, table := range []string{"file_source", "file_tag", "file_tag_value", "file_storage"} {
		var count int
		if err := database.db.QueryRowContext(
			t.Context(),
			"SELECT COUNT(*) FROM "+table, //nolint:gosec // table comes from the fixed test allowlist.
		).Scan(&count); err != nil {
			t.Fatalf("count %s rows: %v", table, err)
		}
		if count != 0 {
			t.Errorf("%s row count = %d, want 0 after cascade", table, count)
		}
	}
}

func TestCollectionStateMigrationRollsBack(t *testing.T) {
	database := openTestDatabase(t)
	if _, err := database.db.ExecContext(t.Context(), initialMigration); err != nil {
		t.Fatalf("apply migration 1: %v", err)
	}
	if _, err := database.db.ExecContext(
		t.Context(),
		"CREATE TABLE file_source (conflict TEXT)",
	); err != nil {
		t.Fatalf("create conflicting table: %v", err)
	}
	if err := database.Initialize(t.Context()); err == nil {
		t.Fatal("Initialize() error = nil, want migration failure")
	}
	version, err := schemaVersion(t.Context(), database.db)
	if err != nil {
		t.Fatalf("schemaVersion() error = %v", err)
	}
	if version != 1 {
		t.Fatalf("schemaVersion() after rollback = %d, want 1", version)
	}
	if tableExists(t, database.db, "file") {
		t.Fatal("file table exists after migration rollback")
	}
}

func openTestDatabase(t *testing.T) *Database {
	t.Helper()
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "main.sqlite"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return database
}

func tableExists(t *testing.T, database *sql.DB, name string) bool {
	t.Helper()
	var count int
	if err := database.QueryRowContext(
		t.Context(),
		"SELECT COUNT(*) FROM sqlite_schema WHERE type = 'table' AND name = ?",
		name,
	).Scan(&count); err != nil {
		t.Fatalf("query table %q: %v", name, err)
	}
	return count == 1
}

func assertStatementFails(t *testing.T, database *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := database.ExecContext(t.Context(), query, args...); err == nil {
		t.Fatalf("ExecContext(%q) error = nil, want constraint failure", query)
	}
}
