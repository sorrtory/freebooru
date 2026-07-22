// Package main provides the FreeBooru HTTP server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/sorrtory/freebooru/internal/bootstrap"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	app, err := bootstrap.NewCore(logger)
	if err != nil {
		logger.Error("failed to start FreeBooru", "error", err)
		os.Exit(1)
	}
	if err := app.LoadConfig(context.Background()); err != nil {
		logger.Error("failed to load FreeBooru configuration", "error", err)
		os.Exit(1)
	}

	for _, diagnostic := range app.CheckConfig(context.Background()) {
		logger.Warn(
			"configuration problem",
			"severity", diagnostic.Severity,
			"code", diagnostic.Code,
			"file", diagnostic.File,
			"document", diagnostic.Document,
			"field", diagnostic.Field,
			"message", diagnostic.Message,
		)
	}

	// TODO: start the HTTP server with app. Domain configuration diagnostics
	// are intentionally non-fatal.
	fmt.Println("FreeBooru server initialized")
}
