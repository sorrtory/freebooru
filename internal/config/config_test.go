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
	want := DefaultAppConfig()
	want.HTTPPort = 9090
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
	content := `lang: en
default_collection: main
default_storage_name: default
default_storage_path: $HOME/.local/share/freebooru/storage/default
http_port: 70000
remove_on_upload: false
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := (YAMLFile[AppConfig]{Path: path, Default: DefaultAppConfig, Validate: VerifyAppConfig}).Read()
	if err == nil || !strings.Contains(err.Error(), "http_port") {
		t.Fatalf("Read() error = %v, want port validation error", err)
	}
}

func TestLoadAppAppliesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "freebooru.yaml")
	if err := os.WriteFile(path, []byte("remove_on_upload: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadApp(path)
	if err != nil {
		t.Fatalf("LoadApp() error = %v", err)
	}
	want := DefaultAppConfig()
	want.RemoveOnUpload = true
	if got != want {
		t.Fatalf("LoadApp() = %#v, want %#v", got, want)
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
	for _, want := range []string{"path is required", "parse YAML", "tags must contain"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("CheckDomain() error = %q, want %q", err, want)
		}
	}
}

func TestDefaultConfigsRoundTrip(t *testing.T) {
	app := DefaultAppConfig()
	tests := []struct {
		name  string
		write func(string) error
		read  func(string) error
	}{
		{
			name: "application",
			write: func(path string) error {
				return (YAMLFile[AppConfig]{Path: path, Validate: VerifyAppConfig}).Write(app)
			},
			read: func(path string) error {
				_, err := (YAMLFile[AppConfig]{Path: path, Validate: VerifyAppConfig}).Read()
				return err
			},
		},
		{
			name: "storage",
			write: func(path string) error {
				return (YAMLFile[StorageConfig]{Path: path}).Write(DefaultStorageConfig(app))
			},
			read: func(path string) error {
				got, err := (YAMLFile[StorageConfig]{Path: path}).Read()
				if err != nil {
					return err
				}
				for _, provider := range got {
					if err := VerifyStorageProvider(provider); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			name: "collection",
			write: func(path string) error {
				return (YAMLFile[CollectionConfig]{Path: path, Validate: VerifyCollectionConfig}).Write(DefaultCollectionConfig(app))
			},
			read: func(path string) error {
				_, err := (YAMLFile[CollectionConfig]{Path: path, Validate: VerifyCollectionConfig}).Read()
				return err
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.name+".yaml")
			if err := tt.write(path); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			if err := tt.read(path); err != nil {
				t.Fatalf("Read() error = %v", err)
			}
		})
	}
}

func TestYAMLFileRejectsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "freebooru.yaml")
	content := "lang: en\nunknown: true\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := (YAMLFile[AppConfig]{Path: path}).Read()
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("Read() error = %v, want unknown field error", err)
	}
}

func TestCollectionLocationUsesDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	got, err := CollectionLocation(CollectionConfig{Name: "main"})
	if err != nil {
		t.Fatalf("CollectionLocation() error = %v", err)
	}
	want := filepath.Join(home, ".local", "share", "freebooru", "collections", "main.sqlite")
	if got != want {
		t.Fatalf("CollectionLocation() = %q, want %q", got, want)
	}
}

func TestEnsureDefaultsDoesNotOverwrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureDefaults(paths); err != nil {
		t.Fatalf("EnsureDefaults() error = %v", err)
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
	if _, err := EnsureDefaults(paths); err != nil {
		t.Fatalf("second EnsureDefaults() error = %v", err)
	}
	got, err := os.ReadFile(paths.App)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Fatalf("EnsureDefaults() overwrote existing app config: %q", got)
	}
}
