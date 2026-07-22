package core

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/sorrtory/freebooru/internal/collection"
)

// ImportResult identifies newly indexed content and its physical assignments.
type ImportResult struct {
	SHA256        string
	SizeBytes     int64
	Storages      []string
	RecordCreated bool
	CreatedCopies []string
}

// Import validates and transactionally indexes one regular local file.
func (c *Core) Import(ctx context.Context, request ImportRequest) (result ImportResult, err error) {
	prepared, err := c.prepareImport(request)
	if err != nil {
		return ImportResult{}, err
	}
	sourcePath, err := filepath.Abs(request.SourcePath)
	if err != nil {
		return ImportResult{}, fmt.Errorf("resolve import source %q: %w", request.SourcePath, err)
	}
	database, err := c.openCollectionDatabase(ctx, prepared.collection.Name)
	if err != nil {
		return ImportResult{}, err
	}
	closed := false
	defer func() {
		if !closed {
			if closeErr := database.Close(); closeErr != nil {
				err = errors.Join(err, fmt.Errorf(
					"close collection %q after import failure: %w",
					prepared.collection.Name,
					closeErr,
				))
			}
		}
	}()
	copies, err := storeImportCopies(ctx, sourcePath, prepared.storages)
	if err != nil {
		return ImportResult{}, errors.Join(err, cleanupStoredCopies(copies))
	}
	first := copies[0].file
	record, err := database.CreateFile(ctx, collection.NewFile{
		SHA256:         first.SHA256,
		SizeBytes:      first.SizeBytes,
		SourcePath:     sourcePath,
		SourceFilename: filepath.Base(sourcePath),
		Tags:           importTagRecords(c.catalog, prepared.values),
		Storages:       storageNamesFromProviders(prepared.storages),
	})
	if err != nil {
		return ImportResult{}, errors.Join(err, cleanupStoredCopies(copies))
	}
	result = ImportResult{
		SHA256:        record.SHA256,
		SizeBytes:     record.SizeBytes,
		Storages:      append([]string(nil), record.Storages...),
		RecordCreated: true,
		CreatedCopies: createdStorageNames(copies),
	}
	if err := database.Close(); err != nil {
		return result, fmt.Errorf("close imported collection %q: %w", prepared.collection.Name, err)
	}
	closed = true
	if c.config.RemoveOnUpload {
		if err := removeImportedSource(ctx, sourcePath, first, copies); err != nil {
			return result, err
		}
	}
	return result, nil
}
