package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ForEachFile visits a stable collection snapshot in SHA-256 order. It loads
// only the current reconstructed file into memory and stops at the first error.
func (d *Database) ForEachFile(
	ctx context.Context,
	visit func(FileRecord) error,
) (err error) {
	if visit == nil {
		return fmt.Errorf("visit collection file callback is nil")
	}
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("begin file iteration transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, fmt.Errorf("rollback file iteration transaction: %w", rollbackErr))
		}
	}()
	previousHash := ""
	for {
		var sha256 string
		err := tx.QueryRowContext(
			ctx,
			"SELECT sha256 FROM file WHERE sha256 > ? ORDER BY sha256 LIMIT 1",
			previousHash,
		).Scan(&sha256)
		if errors.Is(err, sql.ErrNoRows) {
			break
		}
		if err != nil {
			return fmt.Errorf("query next collection file: %w", err)
		}
		file, err := loadFile(ctx, tx, sha256)
		if err != nil {
			return fmt.Errorf("load collection file during iteration: %w", err)
		}
		if err := visit(file); err != nil {
			return fmt.Errorf("visit collection file %q: %w", sha256, err)
		}
		previousHash = sha256
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit file iteration transaction: %w", err)
	}
	return nil
}
