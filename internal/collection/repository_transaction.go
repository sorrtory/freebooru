package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

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
