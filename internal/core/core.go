// Package core provides application level business logic
package core

import (
	"log/slog"
	"os"
	"path/filepath"
)

type Options struct {
	// config paths
	configDir      string
	configFile     string
	storageFile    string
	tagsDir        string
	collectionsDir string
}

type Core struct {
	log *slog.Logger
	opt Options
	// storage Database
	// config  Config
}

func NewCore(logger *slog.Logger, opt Options) *Core {
	if logger == nil {
		panic("logger is nil")
	}
	return &Core{
		log: logger,
		opt: opt,
	}
}

func NewDefaultCore() *Core {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Set up pathes for config dir
	configDirOS, err := os.UserConfigDir()
	if err != nil {
		logger.Error("Failed to get user config dir", "error", err)
		os.Exit(1)
	}
	configDir := filepath.Join(configDirOS, "freebooru")
	
	// Create the core application
	app := NewCore(logger, Options{
		configDir:      configDir,
		configFile:     filepath.Join(configDir, "freebooru.yaml"),
		storageFile:    filepath.Join(configDir, "storage.yaml"),
		tagsDir:        filepath.Join(configDir, "tags"),
		collectionsDir: filepath.Join(configDir, "collections"),
	})

	return app
}

func (c *Core) Fatal(err error, msg string) {
	c.log.Error(msg, "error", err)
	os.Exit(1)
}
