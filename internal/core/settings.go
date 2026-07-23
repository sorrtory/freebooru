package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/sorrtory/freebooru/internal/config"
)

// Settings is the safe remotely editable application configuration subset.
type Settings struct {
	Revision           string
	Language           string
	DefaultCollection  string
	DefaultStorageName string
	HTTPAddress        string
	HTTPPort           int
	RemoveOnUpload     bool
	Collections        []string
	Storages           []string
}

// SettingsUpdate replaces every safe editable application setting.
type SettingsUpdate struct {
	ExpectedRevision   string
	Language           string
	DefaultCollection  string
	DefaultStorageName string
	HTTPAddress        string
	HTTPPort           int
	RemoveOnUpload     bool
}

// Settings returns safe application values and an opaque concurrency revision.
func (c *Core) Settings() Settings {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	return c.settingsWithOptions(c.config)
}

// UpdateSettings atomically validates and publishes safe application values.
func (c *Core) UpdateSettings(ctx context.Context, update SettingsUpdate) (Settings, error) {
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	if err := ctx.Err(); err != nil {
		return Settings{}, err
	}
	if c.session != nil {
		return Settings{}, fmt.Errorf("close collection %q before changing settings", c.session.name)
	}
	if update.ExpectedRevision == "" || update.ExpectedRevision != settingsRevision(c.config) {
		return Settings{}, fmt.Errorf("settings changed since they were loaded")
	}
	previous := c.config
	next := previous
	next.Lang = update.Language
	next.DefaultCollection = update.DefaultCollection
	next.DefaultStorageName = update.DefaultStorageName
	next.HTTPAddress = update.HTTPAddress
	next.HTTPPort = update.HTTPPort
	next.RemoveOnUpload = update.RemoveOnUpload
	if err := config.ReplaceApp(c.paths.App, next); err != nil {
		return Settings{}, err
	}
	catalog, diagnostics := config.LoadCatalog(c.paths, next)
	graph, graphDiagnostics := config.BuildValidatedGraph(catalog)
	diagnostics = append(diagnostics, graphDiagnostics...)
	if diagnostics.HasErrors() {
		if rollbackErr := config.ReplaceApp(c.paths.App, previous); rollbackErr != nil {
			return Settings{}, fmt.Errorf("settings are invalid and rollback failed: %w", rollbackErr)
		}
		return Settings{}, fmt.Errorf("settings are invalid: %s", diagnostics[0].Message)
	}
	c.config, c.catalog, c.graph = next, catalog, graph
	return c.settingsWithOptions(next), nil
}

func (c *Core) settingsWithOptions(app config.AppConfig) Settings {
	settings := settingsFromConfig(app)
	if c.catalog == nil {
		return settings
	}
	for _, collectionConfig := range c.catalog.Collections() {
		settings.Collections = append(settings.Collections, collectionConfig.Name)
	}
	for _, storage := range c.catalog.Storages() {
		settings.Storages = append(settings.Storages, storage.Name)
	}
	return settings
}

func settingsFromConfig(app config.AppConfig) Settings {
	return Settings{
		Revision: settingsRevision(app), Language: app.Lang,
		DefaultCollection:  app.DefaultCollection,
		DefaultStorageName: app.DefaultStorageName, HTTPAddress: app.HTTPAddress,
		HTTPPort:       app.HTTPPort,
		RemoveOnUpload: app.RemoveOnUpload,
	}
}

func settingsRevision(app config.AppConfig) string {
	value := fmt.Sprintf(
		"%s\x00%s\x00%s\x00%s\x00%s\x00%d\x00%t",
		app.Lang,
		app.DefaultCollection,
		app.DefaultStorageName,
		app.DefaultStoragePath,
		app.HTTPAddress,
		app.HTTPPort,
		app.RemoveOnUpload,
	)
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
