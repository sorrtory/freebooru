package config

import (
	"errors"
	"io"
	"os"

	"github.com/goccy/go-yaml"
)

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
