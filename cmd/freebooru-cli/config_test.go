package main

import (
	"strings"
	"testing"

	appconfig "github.com/sorrtory/freebooru/internal/config"
)

func TestConfigDiagnosticsError(t *testing.T) {
	err := configDiagnosticsError{diagnostics: appconfig.Diagnostics{
		{
			Severity: appconfig.SeverityError,
			Code:     "tag.invalid",
			Message:  "name is required",
			File:     "/config/tags/example.yaml",
			Document: 2,
			Field:    "name",
		},
	}}
	message := err.Error()
	for _, want := range []string{
		"configuration check failed",
		"/config/tags/example.yaml: document 2: name",
		"[tag.invalid] name is required",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("Error() = %q, want %q", message, want)
		}
	}
}
