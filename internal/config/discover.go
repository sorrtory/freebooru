package config

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

func checkYAMLDir[T any](
	root string,
	kind string,
	allowMultiple bool,
	verify func(T) error,
) Diagnostics {
	var diagnostics Diagnostics
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			diagnostics = append(diagnostics, newDiagnostic(
				kind+".walk",
				walkErr.Error(),
				path,
				0,
			))
			return nil
		}
		if entry.IsDir() || !isYAML(path) {
			return nil
		}
		diagnostics = append(
			diagnostics,
			checkYAMLFile(path, kind, allowMultiple, verify)...,
		)
		return nil
	})
	if err != nil {
		diagnostics = append(diagnostics, newDiagnostic(
			kind+".walk",
			fmt.Sprintf("walk directory: %v", err),
			root,
			0,
		))
	}
	return diagnostics
}

func loadYAMLDir[T any](
	root string,
	kind string,
	allowMultiple bool,
	verify func(T) error,
) ([]sourcedDocument[T], Diagnostics) {
	var documents []sourcedDocument[T]
	var diagnostics Diagnostics
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			diagnostics = append(diagnostics, newDiagnostic(kind+".walk", walkErr.Error(), path, 0))
			return nil
		}
		if entry.IsDir() || !isYAML(path) {
			return nil
		}
		loaded, found := loadYAMLFile(path, kind, allowMultiple, verify)
		documents = append(documents, loaded...)
		diagnostics = append(diagnostics, found...)
		return nil
	})
	if err != nil {
		diagnostics = append(diagnostics, newDiagnostic(
			kind+".walk",
			fmt.Sprintf("walk directory: %v", err),
			root,
			0,
		))
	}
	return documents, diagnostics
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
