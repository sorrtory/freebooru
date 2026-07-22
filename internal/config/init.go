package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func Init(paths Paths) error {
	if err := os.MkdirAll(paths.Tags, 0o755); err != nil {
		return fmt.Errorf("create tags directory: %w", err)
	}
	if err := os.MkdirAll(paths.Collections, 0o755); err != nil {
		return fmt.Errorf("create collections directory: %w", err)
	}
	app := DefaultAppConfig()
	if err := writeIfMissing(YAMLFile[AppConfig]{Path: paths.App, Validate: VerifyAppConfig}, app); err != nil {
		return err
	}
	if err := writeIfMissing(YAMLFile[StorageConfig]{Path: paths.Storage}, DefaultStorageConfig(app)); err != nil {
		return err
	}
	return nil
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
