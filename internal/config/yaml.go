package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type YAMLFile[T any] struct {
	Path     string
	Default  func() T
	Validate func(T) error
}

func (f YAMLFile[T]) Read() (T, error) {
	var value T
	if f.Default != nil {
		value = f.Default()
	}
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return value, fmt.Errorf("read %q: %w", f.Path, err)
	}
	if err := yaml.UnmarshalWithOptions(data, &value, yaml.Strict()); err != nil {
		return value, fmt.Errorf("parse %q: %w", f.Path, err)
	}
	if f.Validate != nil {
		if err := f.Validate(value); err != nil {
			return value, fmt.Errorf("verify %q: %w", f.Path, err)
		}
	}
	return value, nil
}

func (f YAMLFile[T]) Write(value T) error {
	if f.Validate != nil {
		if err := f.Validate(value); err != nil {
			return fmt.Errorf("verify %q: %w", f.Path, err)
		}
	}
	data, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal %q: %w", f.Path, err)
	}
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o755); err != nil {
		return fmt.Errorf("create directory for %q: %w", f.Path, err)
	}
	if err := os.WriteFile(f.Path, data, 0o600); err != nil {
		return fmt.Errorf("write %q: %w", f.Path, err)
	}
	return nil
}
