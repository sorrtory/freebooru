package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYAMLFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "freebooru.yaml")
	file := YAMLFile[AppConfig]{Path: path, Validate: VerifyAppConfig}
	want := AppConfig{HTTPPort: 9090}
	if err := file.Write(want); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	got, err := file.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != want {
		t.Fatalf("Read() = %#v, want %#v", got, want)
	}
}

func TestYAMLFileRejectsInvalidAppConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "freebooru.yaml")
	if err := os.WriteFile(path, []byte("http_port: 70000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := (YAMLFile[AppConfig]{Path: path, Validate: VerifyAppConfig}).Read()
	if err == nil || !strings.Contains(err.Error(), "http_port") {
		t.Fatalf("Read() error = %v, want port validation error", err)
	}
}

func TestCheckYAMLDirLoadsNestedMultipleDocuments(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "touhou")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "name: cirno\ntype: multivalue\n---\nname: reimu\n"
	if err := os.WriteFile(filepath.Join(nested, "characters.yaml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	err := checkYAMLDir(root, VerifyTagConfig)
	if err == nil || !strings.Contains(err.Error(), "document 2") {
		t.Fatalf("checkYAMLDir() error = %v, want document 2 error", err)
	}
}

func TestCheckDomainAggregatesBrokenConfigs(t *testing.T) {
	paths, err := PathsFromDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Tags, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Collections, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.Storage, []byte("- name: local\n  type: local\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.Tags, "broken.yaml"), []byte("name: [\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.Collections, "main.yaml"), []byte("name: main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err = CheckDomain(paths)
	if err == nil {
		t.Fatal("CheckDomain() succeeded, want errors")
	}
	for _, want := range []string{"path is required", "parse YAML", "location is required"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("CheckDomain() error = %q, want %q", err, want)
		}
	}
}

func TestInitCreatesLayoutWithoutOverwriting(t *testing.T) {
	paths, err := PathsFromDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := Init(paths); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	for _, path := range []string{paths.App, paths.Storage, paths.Tags, paths.Collections} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %q: %v", path, err)
		}
	}
	custom := []byte("http_port: 1234\n")
	if err := os.WriteFile(paths.App, custom, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Init(paths); err != nil {
		t.Fatalf("second Init() error = %v", err)
	}
	got, err := os.ReadFile(paths.App)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Fatalf("Init() overwrote existing app config: %q", got)
	}
}
