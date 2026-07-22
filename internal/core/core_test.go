package core

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

type fakeDatabase struct {
	initializeErr error
	closeErr      error
	initialized   bool
	closed        bool
	files         []collection.FileRecord
}

func (d *fakeDatabase) Initialize(context.Context) error {
	d.initialized = true
	return d.initializeErr
}

func (d *fakeDatabase) Close() error {
	d.closed = true
	return d.closeErr
}

func (d *fakeDatabase) ForEachFile(
	_ context.Context,
	visit func(collection.FileRecord) error,
) error {
	for _, file := range d.files {
		if err := visit(file); err != nil {
			return err
		}
	}
	return nil
}

func TestInitInitializesAndClosesDefaultCollection(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := config.PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	database := &fakeDatabase{}
	var openedPath string
	app, err := New(testLogger(), paths, func(_ context.Context, path string) (CollectionDatabase, error) {
		openedPath = path
		return database, nil
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := app.Init(t.Context()); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	wantPath := filepath.Join(home, ".local", "share", "freebooru", "collections", "main.sqlite")
	if openedPath != wantPath {
		t.Fatalf("opened path = %q, want %q", openedPath, wantPath)
	}
	if !database.initialized || !database.closed {
		t.Fatalf("database lifecycle: initialized=%t closed=%t", database.initialized, database.closed)
	}
	if app.AppConfig() != config.DefaultAppConfig() {
		t.Fatalf("AppConfig() = %#v", app.AppConfig())
	}
}

func TestInitClosesDatabaseAfterInitializeFailure(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := config.PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	database := &fakeDatabase{initializeErr: errors.New("broken schema")}
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		return database, nil
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	err = app.Init(t.Context())
	if err == nil || !strings.Contains(err.Error(), "broken schema") {
		t.Fatalf("Init() error = %v, want schema error", err)
	}
	if !database.closed {
		t.Fatal("Init() did not close database after initialization failure")
	}
}

func TestInitReturnsOpenFailureAndCanRetry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := config.PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	database := &fakeDatabase{}
	attempt := 0
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		attempt++
		if attempt == 1 {
			return nil, errors.New("database unavailable")
		}
		return database, nil
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := app.Init(t.Context()); err == nil || !strings.Contains(err.Error(), "database unavailable") {
		t.Fatalf("first Init() error = %v", err)
	}
	if err := app.Init(t.Context()); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}
	if !database.initialized || !database.closed {
		t.Fatalf("database lifecycle after retry: initialized=%t closed=%t", database.initialized, database.closed)
	}
}

func TestCheckConfigBuildsCatalogAndGraphDiagnostics(t *testing.T) {
	paths, appConfig := provisionTestConfig(t)
	tag := "name: artist\ntype: text\ndemand:\n  - tag: missing\n"
	if err := os.WriteFile(filepath.Join(paths.Tags, "artist.yaml"), []byte(tag), 0o600); err != nil {
		t.Fatal(err)
	}
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		return &fakeDatabase{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	if app.AppConfig() != appConfig {
		t.Fatalf("AppConfig() = %#v, want %#v", app.AppConfig(), appConfig)
	}
	diagnostics := app.CheckConfig(t.Context())
	if !hasDiagnosticCode(diagnostics, "relationship.target_missing") {
		t.Fatalf("CheckConfig() diagnostics = %#v", diagnostics)
	}
	tags, err := app.SearchTags("AR")
	if err != nil {
		t.Fatalf("SearchTags() error = %v", err)
	}
	if len(tags) != 1 || tags[0].Name != "artist" {
		t.Fatalf("SearchTags(AR) = %#v", tags)
	}
}

func TestOpenCollectionRejectsInvalidAndOpensIndependentValidCollection(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	invalid := "name: broken\ntags:\n  require:\n    - storage: missing\n"
	if err := os.WriteFile(
		filepath.Join(paths.Collections, "broken.yaml"),
		[]byte(invalid),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	database := &fakeDatabase{}
	openCalls := 0
	app, err := New(testLogger(), paths, func(context.Context, string) (CollectionDatabase, error) {
		openCalls++
		return database, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := app.LoadConfig(t.Context()); err != nil {
		t.Fatal(err)
	}
	diagnostics := app.CheckConfig(t.Context())
	if !diagnostics.HasErrors() {
		t.Fatal("CheckConfig() diagnostics has no errors")
	}
	if _, err := app.OpenCollection(t.Context(), "broken"); err == nil {
		t.Fatal("OpenCollection(broken) error = nil")
	}
	if openCalls != 0 {
		t.Fatalf("opener called %d times for invalid collection", openCalls)
	}
	opened, err := app.OpenCollection(t.Context(), "MAIN")
	if err != nil {
		t.Fatalf("OpenCollection(MAIN) error = %v", err)
	}
	if opened != database || !database.initialized || openCalls != 1 {
		t.Fatalf("opened=%#v initialized=%t calls=%d", opened, database.initialized, openCalls)
	}
}

func TestOpenCollectionClosesAfterInitializeFailure(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	database := &fakeDatabase{initializeErr: errors.New("broken schema")}
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
	if _, err := app.OpenCollection(t.Context(), "main"); err == nil {
		t.Fatal("OpenCollection(main) error = nil")
	}
	if !database.closed {
		t.Fatal("OpenCollection() did not close database after initialize failure")
	}
}

func TestOpenCollectionRejectsIncompatiblePersistedState(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	hash := strings.Repeat("a", 64)
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256:   hash,
		Tags:     []collection.TagRecord{{Name: "missing", Type: "bool"}},
		Storages: []string{"default"},
	}}}
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
	_, err = app.OpenCollection(t.Context(), "main")
	if err == nil || !strings.Contains(err.Error(), hash) ||
		!strings.Contains(err.Error(), "not imported") {
		t.Fatalf("OpenCollection() error = %v, want incompatible file context", err)
	}
	if !database.closed {
		t.Fatal("OpenCollection() did not close incompatible database")
	}
}

func TestOpenCollectionRejectsMissingRequiredPersistedStorage(t *testing.T) {
	paths, _ := provisionTestConfig(t)
	database := &fakeDatabase{files: []collection.FileRecord{{
		SHA256: strings.Repeat("b", 64),
	}}}
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
	_, err = app.OpenCollection(t.Context(), "main")
	if err == nil || !strings.Contains(err.Error(), "required storage") {
		t.Fatalf("OpenCollection() error = %v, want required storage error", err)
	}
	if !database.closed {
		t.Fatal("OpenCollection() did not close database missing required storage")
	}
}

func provisionTestConfig(t *testing.T) (config.Paths, config.AppConfig) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := config.PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	appConfig, err := config.EnsureDefaults(paths)
	if err != nil {
		t.Fatal(err)
	}
	return paths, appConfig
}

func hasDiagnosticCode(diagnostics config.Diagnostics, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
