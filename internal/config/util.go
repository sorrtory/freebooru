package config

import (
	"fmt"
	"log/slog"
	"os"
)

func getConfigDir() (string, error) {
	configDirOS, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %v", err)
	}
	return configDirOS, nil
}

func createDirIfNotExists(log *slog.Logger, path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Info("Creating directory", "path", path)
			if err := os.Mkdir(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", path, err)
			}
		} else {
			return fmt.Errorf("failed to stat directory %s: %v", path, err)
		}
	}
	return nil
}

func createFileIfNotExists(log *slog.Logger, path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Info("Creating file", "path", path)
			file, err := os.Create(path)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %v", path, err)
			}
			defer file.Close()
		} else {
			return fmt.Errorf("failed to stat file %s: %v", path, err)
		}
	}
	return nil
}
