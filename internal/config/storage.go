package config

import (
	"fmt"
	"strings"
)

type StorageConfig []StorageProvider

type StorageProvider struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

func DefaultStorageConfig(app AppConfig) StorageConfig {
	return StorageConfig{{
		Name: app.DefaultStorageName,
		Type: "local",
		Path: app.DefaultStoragePath,
	}}
}

func VerifyStorageProvider(provider StorageProvider) error {
	if err := verifyName("name", provider.Name); err != nil {
		return err
	}
	if !strings.EqualFold(provider.Type, "local") {
		return fmt.Errorf("type must be local")
	}
	if strings.TrimSpace(provider.Path) == "" {
		return fmt.Errorf("path is required for local storage")
	}
	if _, err := ExpandPath(provider.Path); err != nil {
		return fmt.Errorf("path: %w", err)
	}
	return nil
}
