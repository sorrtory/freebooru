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

// CheckDomain validates storage, tag, and collection configuration.
func CheckDomain(paths Paths) Diagnostics {
	diagnostics := checkStorage(paths.Storage)
	diagnostics = append(diagnostics, checkYAMLDir(
		paths.Tags,
		"tag",
		true,
		VerifyTagConfig,
	)...)
	diagnostics = append(diagnostics, checkYAMLDir(
		paths.Collections,
		"collection",
		false,
		VerifyCollectionConfig,
	)...)
	return diagnostics
}

func checkStorage(path string) Diagnostics {
	storage, err := (YAMLFile[StorageConfig]{Path: path}).Read()
	if err != nil {
		return Diagnostics{newDiagnostic("storage.load", err.Error(), path, 1)}
	}
	var diagnostics Diagnostics
	if len(storage) == 0 {
		diagnostics = append(diagnostics, newDiagnostic(
			"storage.empty",
			"at least one storage provider is required",
			path,
			1,
		))
	}
	for index, provider := range storage {
		if err := VerifyStorageProvider(provider); err != nil {
			diagnostic := newDiagnostic(
				"storage.invalid",
				err.Error(),
				path,
				1,
			)
			diagnostic.Field = fmt.Sprintf("[%d]", index)
			if field := validationField(err); field != "" {
				diagnostic.Field += "." + field
			}
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	return diagnostics
}

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

func checkYAMLFile[T any](
	path string,
	kind string,
	allowMultiple bool,
	verify func(T) error,
) Diagnostics {
	file, err := os.Open(path)
	if err != nil {
		return Diagnostics{newDiagnostic(kind+".open", err.Error(), path, 0)}
	}
	defer func() {
		_ = file.Close()
	}()

	var diagnostics Diagnostics
	decoder := yaml.NewDecoder(file, yaml.Strict())
	document := 0
	for {
		var value T
		err := decoder.Decode(&value)
		if errors.Is(err, io.EOF) {
			break
		}
		document++
		if err != nil {
			return append(diagnostics, newDiagnostic(
				"yaml.decode",
				err.Error(),
				path,
				document,
			))
		}
		if document > 1 && !allowMultiple {
			diagnostics = append(diagnostics, newDiagnostic(
				"yaml.multiple_documents",
				"only one YAML document is allowed",
				path,
				document,
			))
			continue
		}
		if err := verify(value); err != nil {
			diagnostic := newDiagnostic(
				kind+".invalid",
				err.Error(),
				path,
				document,
			)
			diagnostic.Field = validationField(err)
			diagnostics = append(diagnostics, diagnostic)
		}
	}
	if document == 0 {
		diagnostics = append(diagnostics, newDiagnostic(
			"yaml.empty",
			"YAML file is empty",
			path,
			0,
		))
	}
	return diagnostics
}

func validationField(err error) string {
	message := err.Error()
	field, _, found := strings.Cut(message, " ")
	if !found {
		return ""
	}
	return field
}

func newDiagnostic(code, message, file string, document int) Diagnostic {
	return Diagnostic{
		Severity: SeverityError,
		Code:     code,
		Message:  message,
		File:     file,
		Document: document,
	}
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
