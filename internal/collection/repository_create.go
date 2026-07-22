package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ncruces/go-sqlite3"
)

var errDuplicateContent = errors.New("duplicate content identity")

// CreateFile records initial metadata, tags, and storages atomically.
func (d *Database) CreateFile(ctx context.Context, input NewFile) (FileRecord, error) {
	if len(input.Storages) == 0 {
		return FileRecord{}, fmt.Errorf("create file %q: at least one storage is required", input.SHA256)
	}
	err := d.writeTransaction(ctx, "create file", func(tx *sql.Tx) error {
		fileID, err := insertFile(ctx, tx, input)
		if err != nil {
			return err
		}
		if err := insertSource(ctx, tx, fileID, input); err != nil {
			return err
		}
		for _, tag := range input.Tags {
			if err := insertTag(ctx, tx, fileID, tag); err != nil {
				return err
			}
		}
		for _, storage := range input.Storages {
			if _, err := tx.ExecContext(
				ctx,
				"INSERT INTO file_storage (file_id, storage_name) VALUES (?, ?)",
				fileID,
				storage,
			); err != nil {
				return fmt.Errorf("insert storage %q: %w", storage, err)
			}
		}
		return nil
	})
	if errors.Is(err, errDuplicateContent) {
		existing, findErr := d.File(ctx, input.SHA256)
		if findErr != nil {
			return FileRecord{}, errors.Join(err, fmt.Errorf("load duplicate file: %w", findErr))
		}
		return FileRecord{}, &DuplicateFileError{Existing: existing}
	}
	if err != nil {
		return FileRecord{}, err
	}
	file, err := d.File(ctx, input.SHA256)
	if err != nil {
		return FileRecord{}, fmt.Errorf("load created file %q: %w", input.SHA256, err)
	}
	return file, nil
}

func insertFile(ctx context.Context, tx *sql.Tx, input NewFile) (int64, error) {
	result, err := tx.ExecContext(
		ctx,
		"INSERT INTO file (sha256, size_bytes) VALUES (?, ?)",
		input.SHA256,
		input.SizeBytes,
	)
	if err != nil {
		if errors.Is(err, sqlite3.CONSTRAINT_UNIQUE) {
			return 0, fmt.Errorf("%w: %s", errDuplicateContent, input.SHA256)
		}
		return 0, fmt.Errorf("insert file %q: %w", input.SHA256, err)
	}
	fileID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("read inserted file id: %w", err)
	}
	return fileID, nil
}

func insertSource(ctx context.Context, tx *sql.Tx, fileID int64, input NewFile) error {
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO file_source (file_id, source_path, filename)
		 VALUES (?, ?, ?)`,
		fileID,
		input.SourcePath,
		input.SourceFilename,
	); err != nil {
		return fmt.Errorf("insert source for file %q: %w", input.SHA256, err)
	}
	return nil
}
