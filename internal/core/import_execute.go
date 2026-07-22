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
	session, release, err := c.acquireCollectionSession(ctx, prepared.collection.Name)
	if err != nil {
		return ImportResult{}, err
	}
	defer release(&err)
	copies, err := storeImportCopies(ctx, sourcePath, prepared.storages)
	if err != nil {
		return ImportResult{}, errors.Join(err, cleanupStoredCopies(copies))
	}
	first := copies[0].file
	sourceFilename := filepath.Base(sourcePath)
	if request.SourceFilename != "" {
		sourceFilename = filepath.Base(request.SourceFilename)
	}
	record, err := session.database.CreateFile(ctx, collection.NewFile{
		SHA256:         first.SHA256,
		SizeBytes:      first.SizeBytes,
		SourcePath:     sourcePath,
		SourceFilename: sourceFilename,
		Tags:           importTagRecords(c.catalog, prepared.values),
		Storages:       storageNamesFromProviders(prepared.storages),
	})
	if err != nil {
		// A concurrent importer may commit after this operation created the
		// shared content path. The uniqueness loser cannot safely delete that
		// path because the winner may already reference it. An unreferenced copy
		// is permitted safe garbage; a referenced missing copy is not.
		if errors.Is(err, collection.ErrDuplicateFile) {
			return ImportResult{SHA256: first.SHA256, SizeBytes: first.SizeBytes}, err
		}
		return ImportResult{}, errors.Join(err, cleanupStoredCopies(copies))
	}
	result = ImportResult{
		SHA256:        record.SHA256,
		SizeBytes:     record.SizeBytes,
		Storages:      append([]string(nil), record.Storages...),
		RecordCreated: true,
		CreatedCopies: createdStorageNames(copies),
	}
	if c.config.RemoveOnUpload {
		if err := removeImportedSource(ctx, sourcePath, first, copies); err != nil {
			return result, err
		}
	}
	return result, nil
}
