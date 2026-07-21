package main

import (
	"context"
	"os"
	
	"github.com/sorrtory/freebooru/internal/core"
	"github.com/spf13/cobra"
	
	"log/slog"
)

type coreContextKey struct{}

func withCore(ctx context.Context, app *core.Core) context.Context {
	return context.WithValue(ctx, coreContextKey{}, app)
}

func coreFrom(cmd *cobra.Command) *core.Core {
	app, ok := cmd.Context().Value(coreContextKey{}).(*core.Core)
	if !ok {
		panic("core is not initialized")
	}
	return app
}

func createCore() *core.Core {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create the core application
	app := core.NewCore(logger)

	return app
}

func main() {
	Execute()
}
