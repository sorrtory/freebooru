// Package config loads and validates FreeBooru configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

// AppConfig contains process-wide application settings.
type AppConfig struct {
	Lang               string `yaml:"lang"`
	DefaultCollection  string `yaml:"default_collection"`
	DefaultStorageName string `yaml:"default_storage_name"`
	DefaultStoragePath string `yaml:"default_storage_path"`
	HTTPPort           int    `yaml:"http_port"`
	RemoveOnUpload     bool   `yaml:"remove_on_upload"`
}

// DefaultAppConfig returns the application defaults.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Lang:               "en",
		DefaultCollection:  "main",
		DefaultStorageName: "default",
		DefaultStoragePath: "$HOME/.local/share/freebooru/storage/default",
		HTTPPort:           52800,
		RemoveOnUpload:     false,
	}
}

// VerifyAppConfig checks application configuration fields.
func VerifyAppConfig(cfg AppConfig) error {
	if cfg.Lang != "en" && cfg.Lang != "ru" {
		return fmt.Errorf("lang must be en or ru")
	}
	if err := verifyName("default_collection", cfg.DefaultCollection); err != nil {
		return err
	}
	if err := verifyName("default_storage_name", cfg.DefaultStorageName); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.DefaultStoragePath) == "" {
		return fmt.Errorf("default_storage_path is required")
	}
	if _, err := ExpandPath(cfg.DefaultStoragePath); err != nil {
		return fmt.Errorf("default_storage_path: %w", err)
	}
	if cfg.HTTPPort < 1 || cfg.HTTPPort > 65535 {
		return fmt.Errorf("http_port must be between 1 and 65535")
	}
	return nil
}

// LoadApp reads application configuration and applies missing defaults.
func LoadApp(path string) (AppConfig, error) {
	return (YAMLFile[AppConfig]{
		Path:     path,
		Default:  DefaultAppConfig,
		Validate: VerifyAppConfig,
	}).Read()
}

// ReplaceApp atomically replaces a validated application configuration.
func ReplaceApp(path string, app AppConfig) error {
	if err := VerifyAppConfig(app); err != nil {
		return err
	}
	data, err := yaml.Marshal(app)
	if err != nil {
		return fmt.Errorf("marshal application configuration: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".freebooru-app-update-*")
	if err != nil {
		return fmt.Errorf("create temporary application configuration: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("protect temporary application configuration: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary application configuration: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary application configuration: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary application configuration: %w", err)
	}
	if err := os.Rename(temporaryName, path); err != nil {
		return fmt.Errorf("replace application configuration: %w", err)
	}
	return nil
}
