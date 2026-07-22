package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// EnsureDefaults creates missing default configuration and data directories
// without modifying existing YAML files.
func EnsureDefaults(paths Paths) (AppConfig, error) {
	if err := writeIfMissing(
		YAMLFile[AppConfig]{Path: paths.App, Validate: VerifyAppConfig},
		DefaultAppConfig(),
	); err != nil {
		return AppConfig{}, fmt.Errorf("ensure application config: %w", err)
	}
	app, err := LoadApp(paths.App)
	if err != nil {
		return AppConfig{}, err
	}
	if err := os.MkdirAll(paths.Tags, 0o755); err != nil {
		return AppConfig{}, fmt.Errorf("create tags directory: %w", err)
	}
	if err := os.MkdirAll(paths.Collections, 0o755); err != nil {
		return AppConfig{}, fmt.Errorf("create collections directory: %w", err)
	}
	if err := ensureDefaultStorage(paths.Storage, app); err != nil {
		return AppConfig{}, err
	}
	collectionPath := filepath.Join(paths.Collections, app.DefaultCollection+".yaml")
	if err := ensureDefaultCollection(collectionPath, app); err != nil {
		return AppConfig{}, err
	}
	storagePath, err := ExpandPath(app.DefaultStoragePath)
	if err != nil {
		return AppConfig{}, fmt.Errorf("resolve default storage path: %w", err)
	}
	if err := os.MkdirAll(storagePath, 0o755); err != nil {
		return AppConfig{}, fmt.Errorf("create default storage directory %q: %w", storagePath, err)
	}
	collectionLocation, err := CollectionLocation(DefaultCollectionConfig(app))
	if err != nil {
		return AppConfig{}, fmt.Errorf("resolve default collection location: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(collectionLocation), 0o755); err != nil {
		return AppConfig{}, fmt.Errorf("create collections data directory: %w", err)
	}
	return app, nil
}

func ensureDefaultStorage(path string, app AppConfig) error {
	file := YAMLFile[StorageConfig]{Path: path}
	if err := writeIfMissing(file, DefaultStorageConfig(app)); err != nil {
		return fmt.Errorf("ensure storage config: %w", err)
	}
	storage, err := file.Read()
	if err != nil {
		return err
	}
	wantPath, err := ExpandPath(app.DefaultStoragePath)
	if err != nil {
		return err
	}
	for index, provider := range storage {
		if err := VerifyStorageProvider(provider); err != nil {
			return fmt.Errorf("%s: item %d: %w", path, index+1, err)
		}
		if !strings.EqualFold(provider.Name, app.DefaultStorageName) {
			continue
		}
		gotPath, err := ExpandPath(provider.Path)
		if err != nil {
			return fmt.Errorf("%s: item %d: %w", path, index+1, err)
		}
		if !strings.EqualFold(provider.Type, "local") || gotPath != wantPath {
			return fmt.Errorf("default storage %q conflicts with freebooru.yaml", app.DefaultStorageName)
		}
		return nil
	}
	return fmt.Errorf("storage config does not define default storage %q", app.DefaultStorageName)
}

func ensureDefaultCollection(path string, app AppConfig) error {
	want := DefaultCollectionConfig(app)
	file := YAMLFile[CollectionConfig]{Path: path, Validate: VerifyCollectionConfig}
	if err := writeIfMissing(file, want); err != nil {
		return fmt.Errorf("ensure default collection config: %w", err)
	}
	got, err := file.Read()
	if err != nil {
		return err
	}
	gotLocation, err := CollectionLocation(got)
	if err != nil {
		return err
	}
	wantLocation, err := CollectionLocation(want)
	if err != nil {
		return err
	}
	if !strings.EqualFold(got.Name, want.Name) || gotLocation != wantLocation {
		return fmt.Errorf("default collection %q conflicts with freebooru.yaml", app.DefaultCollection)
	}
	for _, ref := range append(got.Tags.Require, got.Tags.Import...) {
		if strings.EqualFold(ref.Storage, app.DefaultStorageName) {
			return nil
		}
	}
	return fmt.Errorf("default collection %q does not reference storage %q", app.DefaultCollection, app.DefaultStorageName)
}

func writeIfMissing[T any](file YAMLFile[T], value T) error {
	_, err := os.Stat(file.Path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat %q: %w", file.Path, err)
	}
	return file.Write(value)
}
