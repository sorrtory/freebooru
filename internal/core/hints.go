package core

import (
	"context"
	"fmt"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

// SearchCollectionTags returns imported tags matching a normalized prefix.
// It does not open SQLite and is suitable for frontend completion.
func (c *Core) SearchCollectionTags(
	collectionName string,
	prefix string,
) ([]config.TagConfig, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.catalog == nil {
		return nil, fmt.Errorf("configuration has not been checked")
	}
	if collectionName == "" {
		collectionName = c.config.DefaultCollection
	}
	var available collectionAvailability
	if c.session != nil {
		if normalizeStateName(c.session.name) != normalizeStateName(collectionName) {
			return nil, fmt.Errorf(
				"collection %q is open; close it before using collection %q",
				c.session.name,
				collectionName,
			)
		}
		available = c.session.availability
	} else {
		references, ok := c.catalog.CollectionReferences(collectionName)
		if !ok {
			return nil, fmt.Errorf("collection %q is missing or invalid", collectionName)
		}
		available = newCollectionAvailability(references)
	}
	candidates := c.catalog.SearchTags(prefix)
	result := make([]config.TagConfig, 0, len(candidates))
	for _, tag := range candidates {
		key := normalizeStateName(tag.Name)
		_, imported := available.importedTags[key]
		if key == "storage" && len(available.importedStorages) > 0 {
			imported = true
		}
		if imported {
			result = append(result, tag)
		}
	}
	return result, nil
}

// AllowedValues evaluates every predefined value against one persisted file.
func (c *Core) AllowedValues(
	ctx context.Context,
	collectionName string,
	sha256 string,
	tagName string,
) (values []evaluator.ValueAvailability, err error) {
	session, release, err := c.acquireCollectionSession(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	defer release(&err)
	key := normalizeStateName(tagName)
	_, imported := session.availability.importedTags[key]
	if key == "storage" && len(session.availability.importedStorages) > 0 {
		imported = true
	}
	if !imported {
		return nil, fmt.Errorf("tag %q is not imported by the collection", tagName)
	}
	record, err := session.database.File(ctx, sha256)
	if err != nil {
		return nil, fmt.Errorf("load file %q for value hints: %w", sha256, err)
	}
	state, err := persistedFileState(record, c.catalog, session.availability)
	if err != nil {
		return nil, fmt.Errorf("reconstruct file %q for value hints: %w", sha256, err)
	}
	values, err = session.evaluator.AllowedValues(tagName, state)
	if err != nil {
		return nil, fmt.Errorf("evaluate values for tag %q: %w", tagName, err)
	}
	return values, nil
}
