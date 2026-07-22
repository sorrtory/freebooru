package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiagnosticsHasErrors(t *testing.T) {
	if (Diagnostics{{Severity: SeverityWarning}}).HasErrors() {
		t.Fatal("warning-only diagnostics reported errors")
	}
	if !(Diagnostics{{Severity: SeverityError}}).HasErrors() {
		t.Fatal("error diagnostic was not reported")
	}
}

func TestCheckYAMLFileReportsEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.yaml")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	diagnostics := checkYAMLFile(path, "tag", true, VerifyTagConfig)
	if len(diagnostics) != 1 || diagnostics[0].Code != "yaml.empty" {
		t.Fatalf("checkYAMLFile() diagnostics = %#v, want yaml.empty", diagnostics)
	}
}

func TestCheckYAMLFileRejectsMultipleCollectionDocuments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collections.yaml")
	content := `name: first
tags:
  require:
    - storage: default
---
name: second
tags:
  require:
    - storage: default
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	diagnostics := checkYAMLFile(path, "collection", false, VerifyCollectionConfig)
	if len(diagnostics) != 1 {
		t.Fatalf("checkYAMLFile() diagnostics = %#v", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Code != "yaml.multiple_documents" || diagnostic.Document != 2 || diagnostic.File != path {
		t.Fatalf("checkYAMLFile() diagnostic = %#v", diagnostic)
	}
}
