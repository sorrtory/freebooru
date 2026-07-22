package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/evaluator"
)

type collectionSession struct {
	name         string
	config       config.CollectionConfig
	references   config.ResolvedReferences
	availability collectionAvailability
	evaluator    *evaluator.Evaluator
	storages     map[string]localStorage
	database     CollectionDatabase
}

// OpenCollection publishes one fully validated Core-owned collection session.
func (c *Core) OpenCollection(ctx context.Context, name string) error {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if name == "" {
		name = c.config.DefaultCollection
	}
	if c.session != nil {
		return fmt.Errorf(
			"collection %q is already open; close it before opening %q",
			c.session.name,
			name,
		)
	}
	session, err := c.buildCollectionSession(ctx, name)
	if err != nil {
		return err
	}
	c.session = session
	return nil
}

type releaseCollectionSession func(*error)

func (c *Core) acquireCollectionSession(
	ctx context.Context,
	requested string,
) (*collectionSession, releaseCollectionSession, error) {
	c.sessionMu.Lock()
	if err := ctx.Err(); err != nil {
		c.sessionMu.Unlock()
		return nil, nil, err
	}
	if requested == "" {
		requested = c.config.DefaultCollection
	}
	if c.session != nil {
		if normalizeStateName(c.session.name) != normalizeStateName(requested) {
			openName := c.session.name
			c.sessionMu.Unlock()
			return nil, nil, fmt.Errorf(
				"collection %q is open; close it before using collection %q",
				openName,
				requested,
			)
		}
		return c.session, func(*error) {
			c.sessionMu.Unlock()
		}, nil
	}
	session, err := c.buildCollectionSession(ctx, requested)
	if err != nil {
		c.sessionMu.Unlock()
		return nil, nil, err
	}
	c.session = session
	released := false
	return session, func(operationErr *error) {
		if released {
			return
		}
		released = true
		c.session = nil
		if closeErr := session.database.Close(); closeErr != nil {
			*operationErr = errors.Join(
				*operationErr,
				fmt.Errorf("close collection %q database: %w", session.name, closeErr),
			)
		}
		c.sessionMu.Unlock()
	}, nil
}

func (s *collectionSession) storage(name string) (localStorage, error) {
	storage, ok := s.storages[normalizeStateName(name)]
	if !ok {
		return localStorage{}, fmt.Errorf("storage %q is not imported by the collection", name)
	}
	return storage, nil
}

func (c *Core) buildCollectionSession(
	ctx context.Context,
	name string,
) (*collectionSession, error) {
	if c.catalog == nil || c.graph == nil {
		return nil, fmt.Errorf("configuration has not been checked")
	}
	collectionConfig, _, ok := c.catalog.Collection(name)
	if !ok {
		return nil, fmt.Errorf("collection %q is missing or invalid", name)
	}
	references, ok := c.catalog.CollectionReferences(collectionConfig.Name)
	if !ok {
		return nil, fmt.Errorf("collection %q references are unavailable", collectionConfig.Name)
	}
	availability := newCollectionAvailability(references)
	checker, err := evaluator.New(c.catalog, c.graph)
	if err != nil {
		return nil, fmt.Errorf("create collection %q evaluator: %w", collectionConfig.Name, err)
	}
	storages := make(map[string]localStorage, len(availability.importedStorages))
	for key := range availability.importedStorages {
		storage, err := c.resolveMutableStorage(key, availability)
		if err != nil {
			return nil, err
		}
		storages[key] = storage
	}
	database, err := c.openCollectionDatabase(ctx, collectionConfig.Name)
	if err != nil {
		return nil, err
	}
	return &collectionSession{
		name:         collectionConfig.Name,
		config:       collectionConfig,
		references:   references,
		availability: availability,
		evaluator:    checker,
		storages:     storages,
		database:     database,
	}, nil
}

// CloseCollection closes and forgets the current session exactly once.
// Repeated calls are no-ops.
func (c *Core) CloseCollection() error {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.session == nil {
		return nil
	}
	session := c.session
	c.session = nil
	if err := session.database.Close(); err != nil {
		return fmt.Errorf("close collection %q database: %w", session.name, err)
	}
	return nil
}

// OpenCollectionName reports the current session without exposing persistence.
func (c *Core) OpenCollectionName() (string, bool) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if c.session == nil {
		return "", false
	}
	return c.session.name, true
}
