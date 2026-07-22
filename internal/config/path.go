package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ExpandPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path is empty")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}
	path = strings.ReplaceAll(path, "$HOME", home)
	if path == "~" {
		path = home
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	} else if strings.HasPrefix(path, "~") {
		return "", fmt.Errorf("unsupported home path %q", path)
	}
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("path %q is relative", path)
	}
	return path, nil
}
