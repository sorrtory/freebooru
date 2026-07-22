package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sorrtory/freebooru/internal/config"
	contentstorage "github.com/sorrtory/freebooru/internal/storage"
)

func TestStoreImportCopiesCleansEveryPartialFailure(t *testing.T) {
	for failAt := 1; failAt <= 3; failAt++ {
		t.Run(fmt.Sprintf("copy_%d", failAt), func(t *testing.T) {
			source, providers := importCopyFixtures(t, 3)
			calls := 0
			copies, err := storeImportCopiesWith(
				t.Context(),
				source,
				providers,
				func(
					ctx context.Context,
					backend *contentstorage.Local,
					path string,
				) (contentstorage.StoredFile, error) {
					calls++
					if calls == failAt {
						return contentstorage.StoredFile{}, errors.New("injected copy failure")
					}
					return backend.Store(ctx, path)
				},
			)
			if err == nil {
				t.Fatal("storeImportCopiesWith() error = nil")
			}
			if cleanupErr := cleanupStoredCopies(copies); cleanupErr != nil {
				t.Fatalf("cleanupStoredCopies() error = %v", cleanupErr)
			}
			assertStorageRootsEmpty(t, providers)
			if _, err := os.Stat(source); err != nil {
				t.Fatalf("source was lost: %v", err)
			}
		})
	}
}

func TestStoreImportCopiesCleansCancellationAtEveryCopy(t *testing.T) {
	for cancelAt := 1; cancelAt <= 3; cancelAt++ {
		t.Run(fmt.Sprintf("copy_%d", cancelAt), func(t *testing.T) {
			source, providers := importCopyFixtures(t, 3)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			calls := 0
			copies, err := storeImportCopiesWith(
				ctx,
				source,
				providers,
				func(
					ctx context.Context,
					backend *contentstorage.Local,
					path string,
				) (contentstorage.StoredFile, error) {
					calls++
					if calls == cancelAt {
						cancel()
					}
					return backend.Store(ctx, path)
				},
			)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("storeImportCopiesWith() error = %v, want context.Canceled", err)
			}
			if cleanupErr := cleanupStoredCopies(copies); cleanupErr != nil {
				t.Fatalf("cleanupStoredCopies() error = %v", cleanupErr)
			}
			assertStorageRootsEmpty(t, providers)
			if _, err := os.Stat(source); err != nil {
				t.Fatalf("source was lost: %v", err)
			}
		})
	}
}

func importCopyFixtures(t *testing.T, count int) (string, []config.StorageProvider) {
	t.Helper()
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, []byte("multi-storage failure fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	providers := make([]config.StorageProvider, 0, count)
	for index := range count {
		providers = append(providers, config.StorageProvider{
			Name: fmt.Sprintf("storage_%d", index),
			Type: "local",
			Path: filepath.Join(t.TempDir(), fmt.Sprintf("storage_%d", index)),
		})
	}
	return source, providers
}

func assertStorageRootsEmpty(t *testing.T, providers []config.StorageProvider) {
	t.Helper()
	for _, provider := range providers {
		entries, err := os.ReadDir(provider.Path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatalf("read storage %q: %v", provider.Name, err)
		}
		if len(entries) != 0 {
			t.Fatalf("storage %q entries = %#v, want empty", provider.Name, entries)
		}
	}
}
