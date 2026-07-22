package config

import (
	"fmt"
	"strings"
)

type AppConfig struct {
	Lang               string `yaml:"lang"`
	DefaultCollection  string `yaml:"default_collection"`
	DefaultStorageName string `yaml:"default_storage_name"`
	DefaultStoragePath string `yaml:"default_storage_path"`
	HTTPPort           int    `yaml:"http_port"`
	RemoveOnUpload     bool   `yaml:"remove_on_upload"`
}

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

func VerifyAppConfig(cfg AppConfig) error {
	if cfg.Lang != "en" && cfg.Lang != "ru" {
		return fmt.Errorf("lang must be en or ru")
	}
	if strings.TrimSpace(cfg.DefaultCollection) == "" {
		return fmt.Errorf("default_collection is required")
	}
	if strings.TrimSpace(cfg.DefaultStorageName) == "" {
		return fmt.Errorf("default_storage_name is required")
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

func LoadApp(path string) (AppConfig, error) {
	return (YAMLFile[AppConfig]{
		Path:     path,
		Default:  DefaultAppConfig,
		Validate: VerifyAppConfig,
	}).Read()
}
