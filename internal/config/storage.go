package config

import (
	"fmt"
	"strings"
)

type StorageConfig []StorageProvider

type StorageProvider struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	Path      string `yaml:"path,omitempty"`
	TokenFile string `yaml:"tokenfile,omitempty"`
}

func DefaultStorageConfig(app AppConfig) StorageConfig {
	return StorageConfig{{
		Name: app.DefaultStorageName,
		Type: "local",
		Path: app.DefaultStoragePath,
	}}
}

func VerifyStorageProvider(provider StorageProvider) error {
	if strings.TrimSpace(provider.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(provider.Type) == "" {
		return fmt.Errorf("type is required")
	}
	if provider.Type == "local" && strings.TrimSpace(provider.Path) == "" {
		return fmt.Errorf("path is required for local storage")
	}
	if provider.Type == "local" {
		if _, err := ExpandPath(provider.Path); err != nil {
			return fmt.Errorf("path: %w", err)
		}
	}
	return nil
}
