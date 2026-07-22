// Package config loads and validates FreeBooru configuration.
package config

import (
	"fmt"
	"strings"
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
