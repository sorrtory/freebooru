// Package collection manages collection SQLite databases.
package collection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ncruces/go-sqlite3"
	sqlite3driver "github.com/ncruces/go-sqlite3/driver"
)

const (
	maxOpenConnections = 4
	busyTimeoutMillis  = 5000
)

// Database is a connection pool for one collection database.
type Database struct {
	db *sql.DB
}

// Option configures how a collection database is opened.
type Option func(*openOptions) error

type openOptions struct{}

// Open opens or creates the SQLite database at an absolute path.
func Open(ctx context.Context, path string, options ...Option) (*Database, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("collection database path %q is relative", path)
	}
	var settings openOptions
	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("collection database option is nil")
		}
		if err := option(&settings); err != nil {
			return nil, fmt.Errorf("configure collection database: %w", err)
		}
	}
	path = filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create collection database directory: %w", err)
	}
	dsn := (&url.URL{
		Scheme: "file",
		Path:   path,
		RawQuery: url.Values{
			"_pragma": {fmt.Sprintf("busy_timeout(%d)", busyTimeoutMillis)},
			"_txlock": {"immediate"},
		}.Encode(),
	}).String()
	db, err := sqlite3driver.Open(dsn, enableForeignKeys)
	if err != nil {
		return nil, fmt.Errorf("open collection database %q: %w", path, err)
	}
	// A small pool permits concurrent readers without multiplying the driver's
	// per-connection Wasm memory cost unnecessarily.
	db.SetMaxOpenConns(maxOpenConnections)
	db.SetMaxIdleConns(maxOpenConnections)
	database := &Database{db: db}
	if err := database.Ping(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, errors.Join(
				err,
				fmt.Errorf("close collection database after ping failure: %w", closeErr),
			)
		}
		return nil, err
	}
	return database, nil
}

func enableForeignKeys(conn *sqlite3.Conn) error {
	enabled, err := conn.Config(sqlite3.DBCONFIG_ENABLE_FKEY, true)
	if err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	if !enabled {
		return fmt.Errorf("enable foreign keys: SQLite left foreign keys disabled")
	}
	return nil
}

// Ping verifies that the collection database is reachable.
func (d *Database) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping collection database: %w", err)
	}
	return nil
}

// Close closes the collection database connection pool.
func (d *Database) Close() error {
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("close collection database: %w", err)
	}
	return nil
}
