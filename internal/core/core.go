// Package core provides application-level business logic.
package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sorrtory/freebooru/internal/collection"
	"github.com/sorrtory/freebooru/internal/config"
)

// CollectionDatabase is the database lifecycle required by Core.
type CollectionDatabase interface {
	Initialize(context.Context) error
	ForEachFile(context.Context, func(collection.FileRecord) error) error
	File(context.Context, string) (collection.FileRecord, error)
	CreateFile(context.Context, collection.NewFile) (collection.FileRecord, error)
	AddTag(context.Context, string, collection.TagRecord) (collection.TagChange, error)
	SetTag(context.Context, string, collection.TagRecord) (collection.TagChange, error)
	RemoveTag(context.Context, string, string) (collection.TagChange, error)
	AddStorage(context.Context, string, string) (collection.StorageChange, error)
	RemoveStorage(context.Context, string, string) (collection.StorageChange, error)
	Search(context.Context, collection.SearchRequest) ([]collection.FileRecord, error)
	Close() error
}

// CollectionOpener opens one collection database.
type CollectionOpener func(context.Context, string) (CollectionDatabase, error)

// Core exposes application workflows to frontends.
type Core struct {
	log     *slog.Logger
	config  config.AppConfig
	paths   config.Paths
	open    CollectionOpener
	catalog *config.Catalog
	graph   *config.Graph
}

// New creates a Core with explicit dependencies.
func New(logger *slog.Logger, paths config.Paths, open CollectionOpener) (*Core, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if open == nil {
		return nil, fmt.Errorf("collection opener is required")
	}
	return &Core{log: logger, paths: paths, open: open}, nil
}

// LoadConfig loads the required application configuration.
func (c *Core) LoadConfig(context.Context) error {
	appConfig, err := config.LoadApp(c.paths.App)
	if err != nil {
		return fmt.Errorf("load application configuration: %w", err)
	}
	c.config = appConfig
	return nil
}

// CheckConfig rebuilds the immutable domain and relationship snapshots and
// returns all discoverable diagnostics.
func (c *Core) CheckConfig(context.Context) config.Diagnostics {
	catalog, diagnostics := config.LoadCatalog(c.paths, c.config)
	graph, graphDiagnostics := config.BuildValidatedGraph(catalog)
	c.catalog = catalog
	c.graph = graph
	return append(diagnostics, graphDiagnostics...)
}

// SearchTags queries the current catalog without filesystem side effects.
func (c *Core) SearchTags(prefix string) ([]config.TagConfig, error) {
	if c.catalog == nil {
		return nil, fmt.Errorf("configuration has not been checked")
	}
	return c.catalog.SearchTags(prefix), nil
}

// OpenCollection opens and initializes one usable catalog collection. Invalid
// collections are absent from the catalog and cannot be opened.
func (c *Core) OpenCollection(ctx context.Context, name string) (CollectionDatabase, error) {
	if c.catalog == nil {
		return nil, fmt.Errorf("configuration has not been checked")
	}
	collectionConfig, _, ok := c.catalog.Collection(name)
	if !ok {
		return nil, fmt.Errorf("collection %q is missing or invalid", name)
	}
	location, err := config.CollectionLocation(collectionConfig)
	if err != nil {
		return nil, fmt.Errorf("resolve collection %q database: %w", collectionConfig.Name, err)
	}
	database, err := c.open(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("open collection %q database: %w", collectionConfig.Name, err)
	}
	if err := database.Initialize(ctx); err != nil {
		return nil, errors.Join(
			fmt.Errorf("initialize collection %q database: %w", collectionConfig.Name, err),
			closeNamedCollection(database, collectionConfig.Name),
		)
	}
	if err := validateCollectionRecords(
		ctx,
		database,
		collectionConfig.Name,
		c.catalog,
		c.graph,
	); err != nil {
		return nil, errors.Join(
			fmt.Errorf("validate collection %q database: %w", collectionConfig.Name, err),
			closeNamedCollection(database, collectionConfig.Name),
		)
	}
	return database, nil
}

// Init provisions the default application layout.
func (c *Core) Init(ctx context.Context) error {
	appConfig, err := config.EnsureDefaults(c.paths)
	if err != nil {
		return err
	}
	location, err := config.CollectionLocation(config.DefaultCollectionConfig(appConfig))
	if err != nil {
		return fmt.Errorf("resolve default collection database: %w", err)
	}
	database, err := c.open(ctx, location)
	if err != nil {
		return fmt.Errorf("open default collection database: %w", err)
	}
	if err := database.Initialize(ctx); err != nil {
		return errors.Join(
			fmt.Errorf("initialize default collection database: %w", err),
			closeCollection(database),
		)
	}
	if err := closeCollection(database); err != nil {
		return err
	}
	c.config = appConfig
	return nil
}

func closeCollection(database CollectionDatabase) error {
	if err := database.Close(); err != nil {
		return fmt.Errorf("close default collection database: %w", err)
	}
	return nil
}

func closeNamedCollection(database CollectionDatabase, name string) error {
	if err := database.Close(); err != nil {
		return fmt.Errorf("close collection %q database: %w", name, err)
	}
	return nil
}

// AppConfig returns the loaded application configuration.
func (c *Core) AppConfig() config.AppConfig { return c.config }
