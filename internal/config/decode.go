package config

import (
	"errors"
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

type sourcedDocument[T any] struct {
	Value  T
	Source Source
}

func checkYAMLFile[T any](
	path string,
	kind string,
	allowMultiple bool,
	verify func(T) error,
) Diagnostics {
	_, diagnostics := loadYAMLFile(path, kind, allowMultiple, verify)
	return diagnostics
}

func loadYAMLFile[T any](
	path string,
	kind string,
	allowMultiple bool,
	verify func(T) error,
) ([]sourcedDocument[T], Diagnostics) {
	file, err := os.Open(path)
	if err != nil {
		return nil, Diagnostics{newDiagnostic(kind+".open", err.Error(), path, 0)}
	}
	defer func() {
		_ = file.Close()
	}()

	var documents []sourcedDocument[T]
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
			return documents, append(diagnostics, newDiagnostic(
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
			continue
		}
		documents = append(documents, sourcedDocument[T]{
			Value:  value,
			Source: Source{File: path, Document: document},
		})
	}
	if document == 0 {
		diagnostics = append(diagnostics, newDiagnostic(
			"yaml.empty",
			"YAML file is empty",
			path,
			0,
		))
	}
	return documents, diagnostics
}
