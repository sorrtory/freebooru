package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
	contentstorage "github.com/sorrtory/freebooru/internal/storage"
)

func TestAddStorageCopiesBeforeLogicalCommit(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	database := storageTestDatabase(stored, []string{"default"})
	archivePath, _ := archiveBackend.ContentPath(stored.SHA256)
	database.onAddStorage = func() {
		if _, err := os.Stat(archivePath); err != nil {
			t.Errorf("archive copy did not exist before logical commit: %v", err)
		}
	}
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	result, err := app.AddStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "ARCHIVE",
	})
	if err != nil {
		t.Fatalf("AddStorage() error = %v", err)
	}
	if !result.Changed || !result.CopyCreated || database.addedStorage != "archive" {
		t.Fatalf("AddStorage() result = %#v, added = %q", result, database.addedStorage)
	}
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("stat archive copy: %v", err)
	}
}

func TestAddStorageCopyFailureDoesNotCommit(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	defaultPath, _ := defaultBackend.ContentPath(stored.SHA256)
	if err := os.Remove(defaultPath); err != nil {
		t.Fatal(err)
	}
	database := storageTestDatabase(stored, []string{"default"})
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	_, err := app.AddStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "archive",
	})
	if err == nil {
		t.Fatal("AddStorage() error = nil, want missing source error")
	}
	if database.addedStorage != "" {
		t.Fatalf("storage committed after copy failure: %q", database.addedStorage)
	}
}

func TestAddStorageRemovesNewCopyAfterDatabaseFailure(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	database := storageTestDatabase(stored, []string{"default"})
	database.addStorageErr = errors.New("database failure")
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	_, err := app.AddStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "archive",
	})
	if err == nil || !strings.Contains(err.Error(), "database failure") {
		t.Fatalf("AddStorage() error = %v", err)
	}
	archivePath, _ := archiveBackend.ContentPath(stored.SHA256)
	if _, err := os.Stat(archivePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("archive copy remains after rollback: %v", err)
	}
}

func TestAddStorageRepeatedCallVerifiesCopyAndReturnsNoChange(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	defaultPath, _ := defaultBackend.ContentPath(stored.SHA256)
	if _, err := archiveBackend.Store(t.Context(), defaultPath); err != nil {
		t.Fatal(err)
	}
	database := storageTestDatabase(stored, []string{"default", "archive"})
	database.addStorageNoChange = true
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	result, err := app.AddStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "archive",
	})
	if err != nil {
		t.Fatalf("AddStorage() error = %v", err)
	}
	if result.Changed || result.CopyCreated {
		t.Fatalf("AddStorage() result = %#v, want verified no-op", result)
	}
}

func TestRemoveStorageCommitsThenDeletesExactlyOneCopy(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	defaultPath, _ := defaultBackend.ContentPath(stored.SHA256)
	if _, err := archiveBackend.Store(t.Context(), defaultPath); err != nil {
		t.Fatal(err)
	}
	database := storageTestDatabase(stored, []string{"default", "archive"})
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	result, err := app.RemoveStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "archive",
	})
	if err != nil {
		t.Fatalf("RemoveStorage() error = %v", err)
	}
	if !result.Changed || !result.CopyDeleted || result.FileDeleted {
		t.Fatalf("RemoveStorage() result = %#v", result)
	}
	archivePath, _ := archiveBackend.ContentPath(stored.SHA256)
	if _, err := os.Stat(archivePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("archive copy still exists: %v", err)
	}
	if _, err := os.Stat(defaultPath); err != nil {
		t.Fatalf("default copy was removed: %v", err)
	}
}

func TestRemoveStorageDatabaseFailurePreservesCopy(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	defaultPath, _ := defaultBackend.ContentPath(stored.SHA256)
	archiveStored, err := archiveBackend.Store(t.Context(), defaultPath)
	if err != nil {
		t.Fatal(err)
	}
	database := storageTestDatabase(stored, []string{"default", "archive"})
	database.removeStorageErr = errors.New("database failure")
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	_, err = app.RemoveStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "archive",
	})
	if err == nil || !strings.Contains(err.Error(), "database failure") {
		t.Fatalf("RemoveStorage() error = %v", err)
	}
	if _, err := os.Stat(archiveStored.ContentPath); err != nil {
		t.Fatalf("copy removed after database failure: %v", err)
	}
}

func TestRemoveStorageRejectsRequiredStorageBeforeCommit(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	database := storageTestDatabase(stored, []string{"default"})
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	_, err := app.RemoveStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "default",
	})
	if err == nil || !strings.Contains(err.Error(), "required storage") {
		t.Fatalf("RemoveStorage() error = %v, want required storage error", err)
	}
	if database.removedStorage != "" {
		t.Fatalf("required storage removal committed: %q", database.removedStorage)
	}
}

func TestRemoveStorageRepeatedCallIsNoOp(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	database := storageTestDatabase(stored, []string{"default"})
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	result, err := app.RemoveStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "archive",
	})
	if err != nil {
		t.Fatalf("RemoveStorage() error = %v", err)
	}
	if result.Changed || result.CopyDeleted || database.removedStorage != "" {
		t.Fatalf("RemoveStorage() result = %#v, removed = %q", result, database.removedStorage)
	}
}

func TestRemoveStorageReportsCommittedRemovalWhenDeleteFails(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	defaultPath, _ := defaultBackend.ContentPath(stored.SHA256)
	archiveStored, err := archiveBackend.Store(t.Context(), defaultPath)
	if err != nil {
		t.Fatal(err)
	}
	database := storageTestDatabase(stored, []string{"default", "archive"})
	database.onRemoveStorage = func() {
		if err := os.Remove(archiveStored.ContentPath); err != nil {
			t.Errorf("remove archive fixture: %v", err)
			return
		}
		if err := os.Mkdir(archiveStored.ContentPath, 0o755); err != nil {
			t.Errorf("replace archive fixture with directory: %v", err)
			return
		}
		if err := os.WriteFile(filepath.Join(archiveStored.ContentPath, "child"), []byte("x"), 0o600); err != nil {
			t.Errorf("write archive fixture child: %v", err)
		}
	}
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, true)

	result, err := app.RemoveStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "archive",
	})
	if err == nil || !strings.Contains(err.Error(), "was removed") {
		t.Fatalf("RemoveStorage() error = %v, want recoverable deletion report", err)
	}
	if !result.Changed || result.CopyDeleted {
		t.Fatalf("RemoveStorage() result = %#v", result)
	}
}

func TestRemoveFinalStorageDeletesRecordAndCopy(t *testing.T) {
	defaultBackend, archiveBackend, stored := seedStorageContent(t)
	defaultPath, _ := defaultBackend.ContentPath(stored.SHA256)
	database := storageTestDatabase(stored, []string{"default"})
	database.removeStorageDeletesFile = true
	app := newStorageMutationTestCore(t, database, defaultBackend, archiveBackend, false)

	result, err := app.RemoveStorage(t.Context(), StorageMutationRequest{
		SHA256: stored.SHA256, Storage: "default",
	})
	if err != nil {
		t.Fatalf("RemoveStorage() error = %v", err)
	}
	if !result.Changed || !result.FileDeleted || !result.CopyDeleted {
		t.Fatalf("RemoveStorage() result = %#v", result)
	}
	if _, err := os.Stat(defaultPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("final copy still exists: %v", err)
	}
}

func storageTestDatabase(
	stored contentstorage.StoredFile,
	storages []string,
) *fakeDatabase {
	rating := "safe"
	return &fakeDatabase{files: []collection.FileRecord{{
		SHA256:    stored.SHA256,
		SizeBytes: stored.SizeBytes,
		Tags: []collection.TagRecord{
			{Name: "reviewed", Type: "bool"},
			{Name: "rating", Type: "value", TextValue: &rating},
		},
		Storages: append([]string(nil), storages...),
	}}}
}

func seedStorageContent(
	t *testing.T,
) (*contentstorage.Local, *contentstorage.Local, contentstorage.StoredFile) {
	t.Helper()
	defaultBackend, err := contentstorage.NewLocal(filepath.Join(t.TempDir(), "default"))
	if err != nil {
		t.Fatal(err)
	}
	archiveBackend, err := contentstorage.NewLocal(filepath.Join(t.TempDir(), "archive"))
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, []byte("freebooru storage mutation"), 0o600); err != nil {
		t.Fatal(err)
	}
	stored, err := defaultBackend.Store(t.Context(), source)
	if err != nil {
		t.Fatal(err)
	}
	return defaultBackend, archiveBackend, stored
}

func newStorageMutationTestCore(
	t *testing.T,
	database CollectionDatabase,
	defaultBackend *contentstorage.Local,
	archiveBackend *contentstorage.Local,
	requireDefault bool,
) *Core {
	t.Helper()
	paths, appConfig := provisionTestConfig(t)
	if err := (config.YAMLFile[config.StorageConfig]{Path: paths.Storage}).Write(config.StorageConfig{
		{Name: "default", Type: "local", Path: storageRoot(t, defaultBackend)},
		{Name: "archive", Type: "local", Path: storageRoot(t, archiveBackend)},
	}); err != nil {
		t.Fatal(err)
	}
	tags := "name: reviewed\ntype: bool\n---\nname: rating\ntype: value\nvalues:\n  - val: safe\n"
	if err := os.WriteFile(filepath.Join(paths.Tags, "storage-test.yaml"), []byte(tags), 0o600); err != nil {
		t.Fatal(err)
	}
	requireStorage := ""
	if requireDefault {
		requireStorage = "    - storage: default\n"
	}
	collectionDocument := "name: main\ntags:\n  require:\n" +
		"    - tag: reviewed\n    - tag: rating\n" + requireStorage +
		"  import:\n    - storage: default\n    - storage: archive\n"
	if err := os.WriteFile(
		filepath.Join(paths.Collections, "main.yaml"),
		[]byte(collectionDocument),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if app.AppConfig() != appConfig {
		t.Fatal("application config changed unexpectedly")
	}
	if diagnostics := app.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	return app
}

func storageRoot(t *testing.T, backend *contentstorage.Local) string {
	t.Helper()
	path, err := backend.ContentPath(strings.Repeat("a", 64))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(filepath.Dir(path))
}
