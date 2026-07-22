package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Search returns complete records matching every term in deterministic order.
func (d *Database) Search(
	ctx context.Context,
	request SearchRequest,
) (files []FileRecord, err error) {
	query, arguments, err := buildSearchSQL(request)
	if err != nil {
		return nil, err
	}
	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin collection search transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, fmt.Errorf("rollback collection search transaction: %w", rollbackErr))
		}
	}()
	rows, err := tx.QueryContext(ctx, query, arguments...)
	if err != nil {
		return nil, fmt.Errorf("query collection search: %w", err)
	}
	var hashes []string
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			return nil, errors.Join(
				fmt.Errorf("scan collection search result: %w", err),
				closeRows(rows, "collection search"),
			)
		}
		hashes = append(hashes, hash)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Join(
			fmt.Errorf("iterate collection search results: %w", err),
			closeRows(rows, "collection search"),
		)
	}
	if err := closeRows(rows, "collection search"); err != nil {
		return nil, err
	}
	files = make([]FileRecord, 0, len(hashes))
	for _, hash := range hashes {
		file, err := loadFile(ctx, tx, hash)
		if err != nil {
			return nil, fmt.Errorf("load collection search result %q: %w", hash, err)
		}
		files = append(files, file)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit collection search transaction: %w", err)
	}
	return files, nil
}
