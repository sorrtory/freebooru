package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalContentPath(t *testing.T) {
	root := t.TempDir()
	backend, err := NewLocal(root)
	if err != nil {
		t.Fatalf("NewLocal() error = %v", err)
	}
	hash := strings.Repeat("ab", sha256.Size)
	got, err := backend.ContentPath(hash)
	if err != nil {
		t.Fatalf("ContentPath() error = %v", err)
	}
	want := filepath.Join(root, "ab", hash)
	if got != want {
		t.Fatalf("ContentPath() = %q, want %q", got, want)
	}
}

func TestLocalContentPathRejectsMalformedHash(t *testing.T) {
	backend, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, hash := range []string{"short", strings.Repeat("A", 64), strings.Repeat("z", 64)} {
		t.Run(hash[:min(len(hash), 8)], func(t *testing.T) {
			if _, err := backend.ContentPath(hash); err == nil {
				t.Fatalf("ContentPath(%q) error = nil", hash)
			}
		})
	}
}

func TestNewLocalRejectsRelativeRoot(t *testing.T) {
	if _, err := NewLocal("relative/storage"); err == nil {
		t.Fatal("NewLocal() error = nil, want relative root error")
	}
}

func TestLocalInspectStreamsFile(t *testing.T) {
	backend, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	contents := []byte(strings.Repeat("freebooru", 32*1024))
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	inspected, err := backend.Inspect(t.Context(), path)
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	wantHash := sha256.Sum256(contents)
	if inspected.Path != path || inspected.SizeBytes != int64(len(contents)) ||
		inspected.SHA256 != hex.EncodeToString(wantHash[:]) {
		t.Fatalf("Inspect() = %#v", inspected)
	}
}

func TestLocalInspectRejectsNonRegularSources(t *testing.T) {
	backend, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if _, err := backend.Inspect(t.Context(), directory); err == nil {
		t.Fatal("Inspect(directory) error = nil")
	}
	target := filepath.Join(directory, "target")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(directory, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := backend.Inspect(t.Context(), link); err == nil {
		t.Fatal("Inspect(symlink) error = nil")
	}
}

func TestLocalInspectHonorsCancellation(t *testing.T) {
	backend, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "source")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = backend.Inspect(ctx, path)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Inspect() error = %v, want context.Canceled", err)
	}
}
