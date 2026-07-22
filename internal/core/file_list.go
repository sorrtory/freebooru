package core

import (
	"context"
	"fmt"
	"sort"

	"github.com/sorrtory/freebooru/internal/collection"
)

// FileStorage identifies one physical storage assignment for a collection file.
type FileStorage struct {
	SHA256  string
	Storage string
}

// ListFiles returns one deterministic row per file/storage assignment.
func (c *Core) ListFiles(
	ctx context.Context,
	collectionName string,
	storageFilter string,
) (rows []FileStorage, err error) {
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	defer release(&err)
	filter := normalizeStateName(storageFilter)
	if filter != "" {
		if _, err := session.storage(storageFilter); err != nil {
			return nil, err
		}
	}
	err = session.database.ForEachFile(ctx, func(file collection.FileRecord) error {
		for _, storage := range file.Storages {
			if filter == "" || normalizeStateName(storage) == filter {
				rows = append(rows, FileStorage{SHA256: file.SHA256, Storage: storage})
			}
		}
		return ctx.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("list collection %q files: %w", session.name, err)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].SHA256 != rows[j].SHA256 {
			return rows[i].SHA256 < rows[j].SHA256
		}
		return normalizeStateName(rows[i].Storage) < normalizeStateName(rows[j].Storage)
	})
	return rows, nil
}
