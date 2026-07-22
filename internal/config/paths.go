package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths contains the standard FreeBooru configuration locations.
type Paths struct {
	Dir         string
	App         string
	Storage     string
	Tags        string
	Collections string
}

// PathsFromDir derives configuration paths from a root directory.
func PathsFromDir(dir string) (Paths, error) {
	if dir == "" {
		return Paths{}, fmt.Errorf("config directory is empty")
	}
	dir = filepath.Clean(dir)
	return Paths{
		Dir:         dir,
		App:         filepath.Join(dir, "freebooru.yaml"),
		Storage:     filepath.Join(dir, "storage.yaml"),
		Tags:        filepath.Join(dir, "tags"),
		Collections: filepath.Join(dir, "collections"),
	}, nil
}

// DefaultPaths returns configuration paths below the user's config directory.
func DefaultPaths() (Paths, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, fmt.Errorf("get user config directory: %w", err)
	}
	return PathsFromDir(filepath.Join(dir, "freebooru"))
}
