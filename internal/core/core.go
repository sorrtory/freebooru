// Package core provides application-level business logic.
package core

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sorrtory/freebooru/internal/config"
)

type Core struct {
	log    *slog.Logger
	config config.AppConfig
	paths  config.Paths
}

func New(logger *slog.Logger, paths config.Paths) (*Core, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	return &Core{log: logger, paths: paths}, nil
}

func (c *Core) LoadConfig(context.Context) error {
	appConfig, err := config.LoadApp(c.paths.App)
	if err != nil {
		return fmt.Errorf("load application configuration: %w", err)
	}
	c.config = appConfig
	return nil
}

func (c *Core) CheckConfig(context.Context) error {
	return config.CheckDomain(c.paths)
}

func (c *Core) InitConfig(context.Context) error {
	return config.Init(c.paths)
}

func (c *Core) AppConfig() config.AppConfig { return c.config }
