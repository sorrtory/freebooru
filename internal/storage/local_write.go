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

// ErrCorruptContent means a derived content path contains unexpected bytes.
var ErrCorruptContent = errors.New("stored content is corrupt")

// StoredFile describes one physical content copy.
type StoredFile struct {
	InspectedFile
	ContentPath string
	Created     bool
}

// Store streams a regular source into this backend and publishes it atomically.
func (s *Local) Store(ctx context.Context, sourcePath string) (result StoredFile, err error) {
	source, err := openRegular(sourcePath)
	if err != nil {
		return StoredFile{}, err
	}
	defer func() {
		if closeErr := source.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close source %q: %w", sourcePath, closeErr))
		}
	}()
	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return StoredFile{}, fmt.Errorf("create local storage root: %w", err)
	}
	staged, err := os.CreateTemp(s.root, ".freebooru-stage-*")
	if err != nil {
		return StoredFile{}, fmt.Errorf("create staged content: %w", err)
	}
	stagedPath := staged.Name()
	stagedClosed := false
	defer func() {
		if !stagedClosed {
			if closeErr := staged.Close(); closeErr != nil {
				err = errors.Join(err, fmt.Errorf("close staged content: %w", closeErr))
			}
		}
		if removeErr := os.Remove(stagedPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("remove staged content: %w", removeErr))
		}
	}()
	digest := sha256.New()
	size, err := io.CopyBuffer(
		io.MultiWriter(staged, digest),
		contextReader{ctx: ctx, reader: source},
		make([]byte, copyBufferSize),
	)
	if err != nil {
		return StoredFile{}, fmt.Errorf("stage source %q: %w", sourcePath, err)
	}
	if err := staged.Sync(); err != nil {
		return StoredFile{}, fmt.Errorf("sync staged content: %w", err)
	}
	if err := staged.Close(); err != nil {
		return StoredFile{}, fmt.Errorf("close staged content: %w", err)
	}
	stagedClosed = true
	hash := hex.EncodeToString(digest.Sum(nil))
	destination, err := s.ContentPath(hash)
	if err != nil {
		return StoredFile{}, err
	}
	stored := StoredFile{
		InspectedFile: InspectedFile{Path: sourcePath, SHA256: hash, SizeBytes: size},
		ContentPath:   destination,
	}
	if _, err := os.Lstat(destination); err == nil {
		if err := verifyStoredContent(ctx, destination, hash, size); err != nil {
			return StoredFile{}, err
		}
		return stored, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return StoredFile{}, fmt.Errorf("inspect destination %q: %w", destination, err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return StoredFile{}, fmt.Errorf("create content shard: %w", err)
	}
	if err := os.Rename(stagedPath, destination); err != nil {
		verifyErr := verifyStoredContent(ctx, destination, hash, size)
		if verifyErr == nil {
			return stored, nil
		}
		return StoredFile{}, errors.Join(
			fmt.Errorf("publish content %q: %w", destination, err),
			verifyErr,
		)
	}
	stored.Created = true
	return stored, nil
}

// Delete removes one content copy. Missing content is an idempotent no-op.
func (s *Local) Delete(hash string) (bool, error) {
	path, err := s.ContentPath(hash)
	if err != nil {
		return false, err
	}
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("delete content %q: %w", path, err)
	}
	// Empty shard cleanup is best effort; another writer may populate it after
	// deletion and before this removal attempt.
	_ = os.Remove(filepath.Dir(path))
	return true, nil
}

func verifyStoredContent(ctx context.Context, path, wantHash string, wantSize int64) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect stored content %q: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return fmt.Errorf("%w: path %q is not a regular file", ErrCorruptContent, path)
	}
	file, err := openRegular(path)
	if err != nil {
		return fmt.Errorf("open stored content %q for verification: %w", path, err)
	}
	digest := sha256.New()
	size, copyErr := io.CopyBuffer(
		digest,
		contextReader{ctx: ctx, reader: file},
		make([]byte, copyBufferSize),
	)
	closeErr := file.Close()
	if copyErr != nil {
		copyErr = fmt.Errorf("verify stored content %q: %w", path, copyErr)
	}
	if err := errors.Join(copyErr, wrapCloseError(closeErr, path)); err != nil {
		return err
	}
	gotHash := hex.EncodeToString(digest.Sum(nil))
	if size != wantSize || gotHash != wantHash {
		return fmt.Errorf(
			"%w: path %q has SHA-256 %s and size %d, want %s and %d",
			ErrCorruptContent,
			path,
			gotHash,
			size,
			wantHash,
			wantSize,
		)
	}
	return nil
}
