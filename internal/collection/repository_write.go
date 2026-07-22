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

// AddStorage adds one logical storage assignment. Repeated adds are no-ops.
func (d *Database) AddStorage(
	ctx context.Context,
	sha256 string,
	storage string,
) (StorageChange, error) {
	var change StorageChange
	err := d.writeTransaction(ctx, "add file storage", func(tx *sql.Tx) error {
		fileID, err := queryFileID(ctx, tx, sha256)
		if err != nil {
			return err
		}
		result, err := tx.ExecContext(
			ctx,
			"INSERT OR IGNORE INTO file_storage (file_id, storage_name) VALUES (?, ?)",
			fileID,
			storage,
		)
		if err != nil {
			return fmt.Errorf("insert storage %q: %w", storage, err)
		}
		changed, err := statementChanged(result)
		if err != nil {
			return fmt.Errorf("read add storage %q result: %w", storage, err)
		}
		change.Changed = changed
		return nil
	})
	if err != nil {
		return StorageChange{}, err
	}
	return change, nil
}

// RemoveStorage removes one logical assignment. Removing the final assignment
// deletes the file record and all dependent rows in the same transaction.
func (d *Database) RemoveStorage(
	ctx context.Context,
	sha256 string,
	storage string,
) (StorageChange, error) {
	var change StorageChange
	err := d.writeTransaction(ctx, "remove file storage", func(tx *sql.Tx) error {
		fileID, err := queryFileID(ctx, tx, sha256)
		if err != nil {
			return err
		}
		var (
			assigned bool
			count    int
		)
		if err := tx.QueryRowContext(
			ctx,
			`SELECT EXISTS(
			     SELECT 1 FROM file_storage
			     WHERE file_id = ? AND storage_name = ?
			 ), COUNT(*)
			 FROM file_storage WHERE file_id = ?`,
			fileID,
			storage,
			fileID,
		).Scan(&assigned, &count); err != nil {
			return fmt.Errorf("query storage %q assignment: %w", storage, err)
		}
		if !assigned {
			return nil
		}
		if count == 1 {
			if _, err := tx.ExecContext(ctx, "DELETE FROM file WHERE file_id = ?", fileID); err != nil {
				return fmt.Errorf("delete file with final storage %q: %w", storage, err)
			}
			change = StorageChange{Changed: true, FileDeleted: true}
			return nil
		}
		if _, err := tx.ExecContext(
			ctx,
			"DELETE FROM file_storage WHERE file_id = ? AND storage_name = ?",
			fileID,
			storage,
		); err != nil {
			return fmt.Errorf("delete storage %q: %w", storage, err)
		}
		change.Changed = true
		return nil
	})
	if err != nil {
		return StorageChange{}, err
	}
	return change, nil
}

func (d *Database) writeTransaction(
	ctx context.Context,
	operation string,
	apply func(*sql.Tx) error,
) (err error) {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin %s transaction: %w", operation, err)
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, fmt.Errorf("rollback %s transaction: %w", operation, rollbackErr))
		}
	}()
	if err := apply(tx); err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit %s transaction: %w", operation, err)
	}
	return nil
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

func insertTag(ctx context.Context, tx *sql.Tx, fileID int64, tag TagRecord) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO file_tag
		    (file_id, tag_name, tag_type, text_value, integer_value)
		 VALUES (?, ?, ?, ?, ?)`,
		fileID,
		tag.Name,
		tag.Type,
		tag.TextValue,
		tag.IntegerValue,
	)
	if err != nil {
		return fmt.Errorf("insert tag %q: %w", tag.Name, err)
	}
	if len(tag.Values) == 0 {
		return nil
	}
	tagID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read inserted tag %q id: %w", tag.Name, err)
	}
	for _, value := range tag.Values {
		if _, err := tx.ExecContext(
			ctx,
			"INSERT INTO file_tag_value (file_tag_id, value) VALUES (?, ?)",
			tagID,
			value,
		); err != nil {
			return fmt.Errorf("insert tag %q value %q: %w", tag.Name, value, err)
		}
	}
	return nil
}

func queryFileID(ctx context.Context, tx *sql.Tx, sha256 string) (int64, error) {
	var fileID int64
	if err := tx.QueryRowContext(
		ctx,
		"SELECT file_id FROM file WHERE sha256 = ?",
		sha256,
	).Scan(&fileID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: %s", ErrFileNotFound, sha256)
		}
		return 0, fmt.Errorf("query file %q id: %w", sha256, err)
	}
	return fileID, nil
}

func statementChanged(result sql.Result) (bool, error) {
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
