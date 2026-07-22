package config

import (
	"fmt"
	"strings"
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

func validationField(err error) string {
	message := err.Error()
	field, _, found := strings.Cut(message, " ")
	if !found {
		return ""
	}
	return field
}
