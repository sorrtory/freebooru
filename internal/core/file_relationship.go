package core

import (
	"context"
	"fmt"

	"github.com/sorrtory/freebooru/internal/collection"
)

type relationshipDatabase interface {
	SetFileParent(context.Context, collection.FileRelationship) error
	RemoveFileParent(context.Context, string) error
	ChildRelationships(context.Context, string) ([]collection.FileRelationship, error)
}

// SetFileParent creates or replaces one child's canonical parent relationship.
func (c *Core) SetFileParent(
	ctx context.Context,
	collectionName string,
	relationship collection.FileRelationship,
) (err error) {
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return err
	}
	defer release(&err)
	database, ok := session.database.(relationshipDatabase)
	if !ok {
		return fmt.Errorf("collection database does not support file relationships")
	}
	if err := database.SetFileParent(ctx, relationship); err != nil {
		return fmt.Errorf("set file parent: %w", err)
	}
	return nil
}

// RemoveFileParent removes a child's relationship without deleting either file.
func (c *Core) RemoveFileParent(
	ctx context.Context,
	collectionName string,
	childSHA256 string,
) (err error) {
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return err
	}
	defer release(&err)
	database, ok := session.database.(relationshipDatabase)
	if !ok {
		return fmt.Errorf("collection database does not support file relationships")
	}
	if err := database.RemoveFileParent(ctx, childSHA256); err != nil {
		return fmt.Errorf("remove file parent: %w", err)
	}
	return nil
}

// FileChildren returns a parent's direct children in deterministic display order.
func (c *Core) FileChildren(
	ctx context.Context,
	collectionName string,
	parentSHA256 string,
) (relationships []collection.FileRelationship, err error) {
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	defer release(&err)
	database, ok := session.database.(relationshipDatabase)
	if !ok {
		return nil, fmt.Errorf("collection database does not support file relationships")
	}
	relationships, err = database.ChildRelationships(ctx, parentSHA256)
	if err != nil {
		return nil, fmt.Errorf("list file children: %w", err)
	}
	return relationships, nil
}
