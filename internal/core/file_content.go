package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileContent is a safe seekable storage copy and its browser metadata.
type FileContent struct {
	SHA256    string
	Filename  string
	MIMEType  string
	SizeBytes int64
	Reader    io.ReadSeekCloser
}

// OpenFileContent opens the first assigned available storage copy.
func (c *Core) OpenFileContent(
	ctx context.Context,
	collectionName string,
	sha256 string,
) (content FileContent, err error) {
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return FileContent{}, err
	}
	released := false
	defer func() {
		if !released {
			release(&err)
		}
	}()
	record, err := session.database.File(ctx, sha256)
	if err != nil {
		return FileContent{}, fmt.Errorf("load file %q: %w", sha256, err)
	}
	if len(record.Storages) == 0 {
		return FileContent{}, fmt.Errorf("file %q has no storage copy", sha256)
	}
	storage, err := session.storage(record.Storages[0])
	if err != nil {
		return FileContent{}, err
	}
	path, err := storage.backend.ContentPath(record.SHA256)
	if err != nil {
		return FileContent{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return FileContent{}, fmt.Errorf("open content for file %q: %w", sha256, err)
	}
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return FileContent{}, errors.Join(fmt.Errorf("inspect content for file %q", sha256), err, file.Close())
	}
	filename := record.SHA256
	if len(record.Sources) > 0 && filepath.Base(record.Sources[len(record.Sources)-1].Filename) != "." {
		filename = filepath.Base(record.Sources[len(record.Sources)-1].Filename)
	}
	release(&err)
	released = true
	if err != nil {
		return FileContent{}, errors.Join(err, file.Close())
	}
	return FileContent{SHA256: record.SHA256, Filename: filename, MIMEType: record.MIMEType, SizeBytes: record.SizeBytes, Reader: file}, nil
}
