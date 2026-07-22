// Package main provides the FreeBooru HTTP server.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sorrtory/freebooru/internal/bootstrap"
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/sorrtory/freebooru/internal/webapi"
	"github.com/sorrtory/freebooru/internal/webui"
)

type configLoader interface {
	LoadConfig(context.Context) error
	AppConfig() config.AppConfig
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(logger); err != nil {
		logger.Error("FreeBooru server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	app, err := bootstrap.NewCore(logger)
	if err != nil {
		return fmt.Errorf("construct application: %w", err)
	}
	port, loadErr := httpPort(context.Background(), app)
	if loadErr != nil {
		logger.Warn(
			"application configuration unavailable; using default HTTP port",
			"port", port,
			"error", loadErr,
		)
	}

	assets, err := webui.Assets()
	if err != nil {
		return fmt.Errorf("load web assets: %w", err)
	}
	api, err := webapi.New(webapi.ModeServer, app)
	if err != nil {
		return fmt.Errorf("construct web API: %w", err)
	}
	handler := webui.NewHandler(api, assets)
	server := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()
	logger.Info("FreeBooru server listening", "address", server.Addr)

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("finish HTTP server: %w", err)
		}
		return nil
	}
}

func httpPort(ctx context.Context, app configLoader) (int, error) {
	if err := app.LoadConfig(ctx); err != nil {
		return config.DefaultAppConfig().HTTPPort, err
	}
	return app.AppConfig().HTTPPort, nil
}
