package core

import (
	"context"
	"fmt"

	"github.com/sorrtory/freebooru/internal/collection"
)

// GetFile loads one indexed file from an explicit or default collection.
func (c *Core) GetFile(
	ctx context.Context,
	collectionName string,
	sha256 string,
) (record collection.FileRecord, err error) {
	if collectionName == "" {
		collectionName = c.config.DefaultCollection
	}
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return collection.FileRecord{}, err
	}
	defer release(&err)

	record, err = session.database.File(ctx, sha256)
	if err != nil {
		return collection.FileRecord{}, fmt.Errorf(
			"load file %q from collection %q: %w",
			sha256,
			collectionName,
			err,
		)
	}
	return record, nil
}
