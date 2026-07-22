package core

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
)

type fakeDatabase struct {
	initializeErr error
	closeErr      error
	initialized   bool
	closed        bool
}

func (d *fakeDatabase) Initialize(context.Context) error {
	d.initialized = true
	return d.initializeErr
}

func (d *fakeDatabase) Close() error {
	d.closed = true
	return d.closeErr
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

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
