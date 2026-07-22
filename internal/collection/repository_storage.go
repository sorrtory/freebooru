package collection

import (
	"context"
	"database/sql"
	"fmt"
)

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
		var assigned bool
		var count int
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
