// Package storage manages physical file content independently of collection
// metadata and tag rules.
package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const copyBufferSize = 128 * 1024

// Local is one content-addressed local storage root.
type Local struct {
	root string
}

// InspectedFile is immutable source metadata calculated from streamed bytes.
type InspectedFile struct {
	Path      string
	SHA256    string
	SizeBytes int64
}

// NewLocal constructs a backend from an already expanded absolute root.
func NewLocal(root string) (*Local, error) {
	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("local storage root %q is relative", root)
	}
	return &Local{root: filepath.Clean(root)}, nil
}

// ContentPath derives the physical path for a validated lowercase SHA-256.
func (s *Local) ContentPath(hash string) (string, error) {
	if err := validateSHA256(hash); err != nil {
		return "", err
	}
	return filepath.Join(s.root, hash[:2], hash), nil
}

// Inspect streams one regular source file to calculate its identity and size.
// It rejects symbolic links and detects replacement between inspection and open.
func (s *Local) Inspect(ctx context.Context, path string) (result InspectedFile, err error) {
	info, err := os.Lstat(path)
	if err != nil {
		return InspectedFile{}, fmt.Errorf("inspect source %q: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return InspectedFile{}, fmt.Errorf("source %q is a symbolic link", path)
	}
	if !info.Mode().IsRegular() {
		return InspectedFile{}, fmt.Errorf("source %q is not a regular file", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return InspectedFile{}, fmt.Errorf("open source %q: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close source %q: %w", path, closeErr))
		}
	}()
	openedInfo, err := file.Stat()
	if err != nil {
		return InspectedFile{}, fmt.Errorf("stat opened source %q: %w", path, err)
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return InspectedFile{}, fmt.Errorf("source %q changed while being opened", path)
	}
	digest := sha256.New()
	size, err := io.CopyBuffer(digest, contextReader{ctx: ctx, reader: file}, make([]byte, copyBufferSize))
	if err != nil {
		return InspectedFile{}, fmt.Errorf("hash source %q: %w", path, err)
	}
	return InspectedFile{
		Path:      path,
		SHA256:    hex.EncodeToString(digest.Sum(nil)),
		SizeBytes: size,
	}, nil
}

func validateSHA256(hash string) error {
	if len(hash) != sha256.Size*2 {
		return fmt.Errorf("SHA-256 %q must contain 64 lowercase hexadecimal characters", hash)
	}
	for _, character := range hash {
		if character < '0' || character > '9' {
			if character < 'a' || character > 'f' {
				return fmt.Errorf("SHA-256 %q must contain 64 lowercase hexadecimal characters", hash)
			}
		}
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(buffer)
	}
}
