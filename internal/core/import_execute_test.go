package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
		SourcePath:     source,
		SourceFilename: "browser-name.txt",
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
	if len(file.Sources) != 1 || file.Sources[0].Path != source || file.Sources[0].Filename != "browser-name.txt" {
		t.Fatalf("persisted sources = %#v", file.Sources)
	}
	duplicate, err := app.Import(t.Context(), ImportRequest{
		SourcePath: source,
		Tags:       map[string]any{"rating": "safe"},
	})
	if !errors.Is(err, collection.ErrDuplicateFile) {
		t.Fatalf("duplicate Import() error = %v, want ErrDuplicateFile", err)
	}
	if duplicate.SHA256 != result.SHA256 || duplicate.SizeBytes != result.SizeBytes {
		t.Fatalf("duplicate Import() result = %#v, want existing content identity", duplicate)
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
	database.onCreateFile = func(input collection.NewFile) {
		root, err := config.ExpandPath(app.config.DefaultStoragePath)
		if err != nil {
			t.Errorf("expand storage path: %v", err)
			return
		}
		if _, err := os.Stat(filepath.Join(root, input.SHA256[:2], input.SHA256)); err != nil {
			t.Errorf("physical copy was not finalized before SQL mutation: %v", err)
		}
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

func TestImportPreservesSourceModifiedAfterCommit(t *testing.T) {
	app := newImportTestCore(t)
	app.config.RemoveOnUpload = true
	database := &fakeDatabase{}
	app.open = func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	}
	source := filepath.Join(t.TempDir(), "modified.txt")
	original := []byte("original import bytes")
	modified := []byte("modified during import")
	if err := os.WriteFile(source, original, 0o600); err != nil {
		t.Fatal(err)
	}
	database.onCreateFile = func(collection.NewFile) {
		if err := os.WriteFile(source, modified, 0o600); err != nil {
			t.Errorf("modify source: %v", err)
		}
	}
	result, err := app.Import(t.Context(), ImportRequest{
		SourcePath: source,
		Tags:       map[string]any{"rating": "safe"},
	})
	if err == nil || !strings.Contains(err.Error(), "source changed after import") {
		t.Fatalf("Import() error = %v, want changed source error", err)
	}
	if database.created == nil || result.SHA256 == "" {
		t.Fatalf("committed import = %#v, result = %#v", database.created, result)
	}
	got, readErr := os.ReadFile(source)
	if readErr != nil {
		t.Fatalf("read preserved source: %v", readErr)
	}
	if string(got) != string(modified) {
		t.Fatalf("source = %q, want modified bytes", got)
	}
	root, expandErr := config.ExpandPath(app.config.DefaultStoragePath)
	if expandErr != nil {
		t.Fatal(expandErr)
	}
	stored, readErr := os.ReadFile(filepath.Join(root, result.SHA256[:2], result.SHA256))
	if readErr != nil {
		t.Fatalf("read committed copy: %v", readErr)
	}
	if string(stored) != string(original) {
		t.Fatalf("stored copy = %q, want original bytes", stored)
	}
}

func TestConcurrentDuplicateImportsKeepWinningCopy(t *testing.T) {
	first := newRealImportTestCore(t)
	second, err := New(testLogger(), first.paths, func(
		ctx context.Context,
		path string,
	) (CollectionDatabase, error) {
		return collection.Open(ctx, path)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := second.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := second.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("second CheckConfig() diagnostics = %#v", diagnostics)
	}
	source := filepath.Join(t.TempDir(), "concurrent.txt")
	contents := []byte("identical concurrent import")
	if err := os.WriteFile(source, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		result ImportResult
		err    error
	}
	start := make(chan struct{})
	outcomes := make(chan outcome, 2)
	var group sync.WaitGroup
	for _, app := range []*Core{first, second} {
		group.Go(func() {
			<-start
			result, err := app.Import(t.Context(), ImportRequest{
				SourcePath: source,
				Tags:       map[string]any{"rating": "safe"},
			})
			outcomes <- outcome{result: result, err: err}
		})
	}
	close(start)
	group.Wait()
	close(outcomes)
	successes := 0
	duplicates := 0
	var hash string
	for found := range outcomes {
		switch {
		case found.err == nil:
			successes++
			hash = found.result.SHA256
		case errors.Is(found.err, collection.ErrDuplicateFile):
			duplicates++
		default:
			t.Fatalf("concurrent Import() error = %v", found.err)
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Fatalf("outcomes: successes = %d, duplicates = %d", successes, duplicates)
	}
	database := openDefaultCollectionDatabase(t, first)
	if _, err := database.File(t.Context(), hash); err != nil {
		t.Fatalf("winning database row: %v", err)
	}
	root, err := config.ExpandPath(first.config.DefaultStoragePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, hash[:2], hash)); err != nil {
		t.Fatalf("winning physical copy: %v", err)
	}
}

func TestRestartIgnoresStageAndAdoptsValidFinalizedOrphan(t *testing.T) {
	first := newRealImportTestCore(t)
	root, err := config.ExpandPath(first.config.DefaultStoragePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	stage := filepath.Join(root, ".freebooru-stage-orphan")
	if err := os.WriteFile(stage, []byte("unfinished"), 0o600); err != nil {
		t.Fatal(err)
	}
	contents := []byte("finalized orphan")
	hashBytes := sha256.Sum256(contents)
	hash := hex.EncodeToString(hashBytes[:])
	finalized := filepath.Join(root, hash[:2], hash)
	if err := os.MkdirAll(filepath.Dir(finalized), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(finalized, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	restarted, err := New(testLogger(), first.paths, func(
		ctx context.Context,
		path string,
	) (CollectionDatabase, error) {
		return collection.Open(ctx, path)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := restarted.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if diagnostics := restarted.CheckConfig(t.Context()); diagnostics.HasErrors() {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := restarted.Import(t.Context(), ImportRequest{
		SourcePath: source,
		Tags:       map[string]any{"rating": "safe"},
	})
	if err != nil {
		t.Fatalf("Import() after restart error = %v", err)
	}
	if result.SHA256 != hash || len(result.CreatedCopies) != 0 {
		t.Fatalf("Import() = %#v, want adopted finalized orphan", result)
	}
	if _, err := os.Stat(stage); err != nil {
		t.Fatalf("stage was not ignored: %v", err)
	}
	if _, err := os.Stat(finalized); err != nil {
		t.Fatalf("finalized orphan disappeared: %v", err)
	}
	database := openDefaultCollectionDatabase(t, restarted)
	if _, err := database.File(t.Context(), hash); err != nil {
		t.Fatalf("adopted database row: %v", err)
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
