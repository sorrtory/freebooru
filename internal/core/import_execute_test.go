package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

func TestImportPersistsContentAndTypedState(t *testing.T) {
	app := newRealImportTestCore(t)
	source := filepath.Join(t.TempDir(), "example.txt")
	contents := []byte("freebooru import")
	if err := os.WriteFile(source, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := app.Import(t.Context(), ImportRequest{
		SourcePath: source,
		Tags: map[string]any{
			"rating": "safe",
			"score":  int64(9),
		},
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	wantHash := sha256.Sum256(contents)
	if result.SHA256 != hex.EncodeToString(wantHash[:]) || result.SizeBytes != int64(len(contents)) {
		t.Fatalf("Import() = %#v", result)
	}
	if !result.RecordCreated || len(result.CreatedCopies) != 1 ||
		result.CreatedCopies[0] != "default" {
		t.Fatalf("Import() creation result = %#v", result)
	}
	storageRoot, err := config.ExpandPath(app.config.DefaultStoragePath)
	if err != nil {
		t.Fatal(err)
	}
	contentPath := filepath.Join(storageRoot, result.SHA256[:2], result.SHA256)
	if _, err := os.Stat(contentPath); err != nil {
		t.Fatalf("stat imported content: %v", err)
	}
	database := openDefaultCollectionDatabase(t, app)
	file, err := database.File(t.Context(), result.SHA256)
	if err != nil {
		t.Fatalf("File() error = %v", err)
	}
	if len(file.Tags) != 3 || len(file.Storages) != 1 || file.Storages[0] != "default" {
		t.Fatalf("persisted file = %#v", file)
	}
	if len(file.Sources) != 1 || file.Sources[0].Path != source {
		t.Fatalf("persisted sources = %#v", file.Sources)
	}
	if _, err := app.Import(t.Context(), ImportRequest{
		SourcePath: source,
		Tags:       map[string]any{"rating": "safe"},
	}); !errors.Is(err, collection.ErrDuplicateFile) {
		t.Fatalf("duplicate Import() error = %v, want ErrDuplicateFile", err)
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatalf("duplicate import changed source: %v", err)
	}
	if _, err := os.Stat(contentPath); err != nil {
		t.Fatalf("duplicate import changed stored content: %v", err)
	}
}

func TestImportRejectsSymlinkBeforeDatabaseMutation(t *testing.T) {
	app := newImportTestCore(t)
	database := &fakeDatabase{}
	app.open = func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	}
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(directory, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Import(t.Context(), ImportRequest{
		SourcePath: link,
		Tags:       map[string]any{"rating": "safe"},
	}); err == nil {
		t.Fatal("Import() error = nil, want symlink rejection")
	}
	if database.created != nil {
		t.Fatalf("database CreateFile() input = %#v, want no mutation", database.created)
	}
}

func TestImportRollsBackNewCopyAfterDatabaseFailure(t *testing.T) {
	app := newImportTestCore(t)
	database := &fakeDatabase{createErr: errors.New("database write failed")}
	app.open = func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	}
	source := filepath.Join(t.TempDir(), "rollback.txt")
	contents := []byte("rollback")
	if err := os.WriteFile(source, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := app.Import(t.Context(), ImportRequest{
		SourcePath: source,
		Tags:       map[string]any{"rating": "safe"},
	})
	if err == nil {
		t.Fatal("Import() error = nil, want database failure")
	}
	hash := sha256.Sum256(contents)
	hashText := hex.EncodeToString(hash[:])
	root, expandErr := config.ExpandPath(app.config.DefaultStoragePath)
	if expandErr != nil {
		t.Fatal(expandErr)
	}
	if _, statErr := os.Stat(filepath.Join(root, hashText[:2], hashText)); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("stored content remains after rollback: %v", statErr)
	}
	if _, statErr := os.Stat(source); statErr != nil {
		t.Fatalf("source changed after rollback: %v", statErr)
	}
}

func TestImportRemoveOnUpload(t *testing.T) {
	app := newRealImportTestCore(t)
	app.config.RemoveOnUpload = true
	source := filepath.Join(t.TempDir(), "move.txt")
	if err := os.WriteFile(source, []byte("move after commit"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := app.Import(t.Context(), ImportRequest{
		SourcePath: source,
		Tags:       map[string]any{"rating": "safe"},
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if _, err := os.Stat(source); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source still exists after remove_on_upload: %v", err)
	}
	database := openDefaultCollectionDatabase(t, app)
	if _, err := database.File(t.Context(), result.SHA256); err != nil {
		t.Fatalf("committed file after source removal: %v", err)
	}
}

func newRealImportTestCore(t *testing.T) *Core {
	t.Helper()
	app := newImportTestCore(t)
	app.open = func(ctx context.Context, path string) (CollectionDatabase, error) {
		return collection.Open(ctx, path)
	}
	return app
}

func openDefaultCollectionDatabase(t *testing.T, app *Core) *collection.Database {
	t.Helper()
	collectionConfig, _, ok := app.catalog.Collection(app.config.DefaultCollection)
	if !ok {
		t.Fatal("default collection is unavailable")
	}
	path, err := config.CollectionLocation(collectionConfig)
	if err != nil {
		t.Fatal(err)
	}
	database, err := collection.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	if err := database.Initialize(t.Context()); err != nil {
		t.Fatal(err)
	}
	return database
}
