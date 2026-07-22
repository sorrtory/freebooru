// Package core provides application-level business logic.
package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/sorrtory/freebooru/internal/config"
)

// CollectionDatabase is the database lifecycle required by Core.
type CollectionDatabase interface {
	Initialize(context.Context) error
	Close() error
}

// OpenCollection opens one collection database.
type OpenCollection func(context.Context, string) (CollectionDatabase, error)

// Core exposes application workflows to frontends.
type Core struct {
	log    *slog.Logger
	config config.AppConfig
	paths  config.Paths
	open   OpenCollection
}

// New creates a Core with explicit dependencies.
func New(logger *slog.Logger, paths config.Paths, open OpenCollection) (*Core, error) {
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

// CheckConfig validates domain configuration.
func (c *Core) CheckConfig(context.Context) error {
	return config.CheckDomain(c.paths)
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

// AppConfig returns the loaded application configuration.
func (c *Core) AppConfig() config.AppConfig { return c.config }
