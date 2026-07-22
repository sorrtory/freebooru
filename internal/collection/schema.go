package collection

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

const currentSchemaVersion = 2

//go:embed migrations/001_initial.sql
var initialMigration string

//go:embed migrations/002_collection_state.sql
var collectionStateMigration string

// Initialize applies supported collection database migrations.
func (d *Database) Initialize(ctx context.Context) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin collection schema transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if _, err := tx.ExecContext(ctx, initialMigration); err != nil {
		return fmt.Errorf("apply collection schema migration 1: %w", err)
	}
	version, err := schemaVersion(ctx, tx)
	if err != nil {
		return err
	}
	if version > currentSchemaVersion {
		return fmt.Errorf(
			"collection schema version %d is newer than supported version %d",
			version,
			currentSchemaVersion,
		)
	}
	if version < 2 {
		if _, err := tx.ExecContext(ctx, collectionStateMigration); err != nil {
			return fmt.Errorf("apply collection schema migration 2: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit collection schema: %w", err)
	}
	return nil
}

func schemaVersion(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
},
) (int, error) {
	var version int
	if err := queryer.QueryRowContext(
		ctx,
		"SELECT COALESCE(MAX(version), 0) FROM schema_migrations",
	).Scan(&version); err != nil {
		return 0, fmt.Errorf("read collection schema version: %w", err)
	}
	return version, nil
}
