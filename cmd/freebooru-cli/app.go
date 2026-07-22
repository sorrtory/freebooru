package main

import (
	"context"
	"log/slog"

	"github.com/sorrtory/freebooru/internal/bootstrap"
	"github.com/sorrtory/freebooru/internal/core"
)

type coreFactory func(*slog.Logger) (*core.Core, error)

func loadCore(ctx context.Context, options *rootOptions) (*core.Core, error) {
	app, err := options.newCore(newLogger(options.verbose))
	if err != nil {
		return nil, err
	}
	if err := app.LoadConfig(ctx); err != nil {
		return nil, err
	}
	diagnostics := app.CheckConfig(ctx)
	if diagnostics.HasErrors() {
		return nil, configDiagnosticsError{diagnostics: diagnostics}
	}
	return app, nil
}

func defaultCoreFactory(logger *slog.Logger) (*core.Core, error) {
	return bootstrap.NewCore(logger)
}
