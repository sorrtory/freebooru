// Package core provides application-level business logic.
package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sorrtory/freebooru/internal/config"
)

// Core exposes application workflows to frontends.
type Core struct {
	log    *slog.Logger
	config config.AppConfig
	paths  config.Paths
}

// New creates a Core with explicit dependencies.
func New(logger *slog.Logger, paths config.Paths) (*Core, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Core{log: logger, paths: paths}, nil
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

// InitConfig provisions the default configuration layout.
func (c *Core) InitConfig(context.Context) error {
	return config.Init(c.paths)
}

// AppConfig returns the loaded application configuration.
func (c *Core) AppConfig() config.AppConfig { return c.config }
