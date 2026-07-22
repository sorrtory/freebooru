package storage

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStoreAndDelete(t *testing.T) {
	root := filepath.Join(t.TempDir(), "storage")
	backend, err := NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(source, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	stored, err := backend.Store(t.Context(), source)
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}
	if !stored.Created || stored.SizeBytes != 0 {
		t.Fatalf("Store() = %#v", stored)
	}
	if _, err := os.Stat(stored.ContentPath); err != nil {
		t.Fatalf("stat stored content: %v", err)
	}
	again, err := backend.Store(t.Context(), source)
	if err != nil {
		t.Fatalf("second Store() error = %v", err)
	}
	if again.Created || again.ContentPath != stored.ContentPath {
		t.Fatalf("second Store() = %#v, want verified no-op", again)
	}
	deleted, err := backend.Delete(stored.SHA256)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !deleted {
		t.Fatal("Delete() = false, want true")
	}
	if _, err := os.Stat(filepath.Dir(stored.ContentPath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("shard directory still exists: %v", err)
	}
	deleted, err = backend.Delete(stored.SHA256)
	if err != nil {
		t.Fatalf("second Delete() error = %v", err)
	}
	if deleted {
		t.Fatal("second Delete() = true, want idempotent no-op")
	}
}

func TestLocalStoreRejectsCorruptExistingContent(t *testing.T) {
	backend, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, []byte("expected"), 0o600); err != nil {
		t.Fatal(err)
	}
	stored, err := backend.Store(t.Context(), source)
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}
	if err := os.WriteFile(stored.ContentPath, []byte("corrupt"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Store(t.Context(), source); !errors.Is(err, ErrCorruptContent) {
		t.Fatalf("second Store() error = %v, want ErrCorruptContent", err)
	}
}

func TestLocalStoreCancellationCleansStage(t *testing.T) {
	root := filepath.Join(t.TempDir(), "storage")
	backend, err := NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := backend.Store(ctx, source); !errors.Is(err, context.Canceled) {
		t.Fatalf("Store() error = %v, want context.Canceled", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read storage root: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("storage root entries = %#v, want no staged content", entries)
	}
}

func TestLocalStoreLargeFile(t *testing.T) {
	backend, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	contents := bytes.Repeat([]byte("freebooru"), 512*1024)
	source := filepath.Join(t.TempDir(), "large")
	if err := os.WriteFile(source, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	stored, err := backend.Store(t.Context(), source)
	if err != nil {
		t.Fatalf("Store() error = %v", err)
	}
	if stored.SizeBytes != int64(len(contents)) {
		t.Fatalf("Store() size = %d, want %d", stored.SizeBytes, len(contents))
	}
	got, err := os.ReadFile(stored.ContentPath)
	if err != nil {
		t.Fatalf("read stored content: %v", err)
	}
	if !bytes.Equal(got, contents) {
		t.Fatal("stored content differs from source")
	}
}

func TestLocalStorePermissionFailureLeavesNoStage(t *testing.T) {
	root := t.TempDir()
	backend, err := NewLocal(root)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(root, 0o700); err != nil {
			t.Errorf("restore storage permissions: %v", err)
		}
	})
	_, storeErr := backend.Store(t.Context(), source)
	if err := os.Chmod(root, 0o700); err != nil {
		t.Fatalf("restore storage permissions: %v", err)
	}
	if storeErr == nil {
		t.Skip("filesystem does not enforce owner write permissions for this process")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read storage root: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("storage root entries = %#v, want no staged content", entries)
	}
}
