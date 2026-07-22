package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

func CheckDomain(paths Paths) error {
	var errs []error

	storage, err := (YAMLFile[StorageConfig]{Path: paths.Storage}).Read()
	if err != nil {
		errs = append(errs, err)
	} else {
		for index, provider := range storage {
			if err := VerifyStorageProvider(provider); err != nil {
				errs = append(errs, fmt.Errorf("%s: item %d: %w", paths.Storage, index+1, err))
			}
		}
	}

	errs = append(errs, checkYAMLDir(paths.Tags, VerifyTagConfig))
	errs = append(errs, checkYAMLDir(paths.Collections, VerifyCollectionConfig))
	return errors.Join(errs...)
}

func checkYAMLDir[T any](root string, verify func(T) error) error {
	var errs []error

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, walkErr))
			return nil
		}
		if entry.IsDir() || !isYAML(path) {
			return nil
		}
		if err := checkYAMLFile(path, verify); err != nil {
			errs = append(errs, err)
		}
		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("walk %q: %w", root, err))
	}
	return errors.Join(errs...)
}

func checkYAMLFile[T any](path string, verify func(T) error) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %q: %w", path, err)
	}
	defer file.Close()

	var errs []error
	decoder := yaml.NewDecoder(file, yaml.Strict())
	for document := 1; ; document++ {
		var value T
		err := decoder.Decode(&value)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("%s: document %d: parse YAML: %w", path, document, err)
		}
		if err := verify(value); err != nil {
			errs = append(errs, fmt.Errorf("%s: document %d: %w", path, document, err))
		}
	}
	return errors.Join(errs...)
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
