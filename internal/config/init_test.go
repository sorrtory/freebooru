package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureDefaultsCreatesLayout(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := PathsFromDir(filepath.Join(home, ".config", "freebooru"))
	if err != nil {
		t.Fatal(err)
	}
	app, err := EnsureStarterDefaults(paths)
	if err != nil {
		t.Fatalf("EnsureDefaults() error = %v", err)
	}
	collectionPath := filepath.Join(paths.Collections, app.DefaultCollection+".yaml")
	storagePath, err := ExpandPath(app.DefaultStoragePath)
	if err != nil {
		t.Fatal(err)
	}
	collectionLocation, err := CollectionLocation(DefaultCollectionConfig(app))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		paths.App,
		paths.Storage,
		paths.Tags,
		paths.Collections,
		collectionPath,
		storagePath,
		filepath.Dir(collectionLocation),
		filepath.Join(paths.Tags, "character.yaml"),
		filepath.Join(paths.Tags, "creator.yaml"),
		filepath.Join(paths.Tags, "general.yaml"),
		filepath.Join(paths.Tags, "metadata.yaml"),
		filepath.Join(paths.Tags, "universe.yaml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %q: %v", path, err)
		}
	}
	collection, err := (YAMLFile[CollectionConfig]{Path: collectionPath}).Read()
	if err != nil {
		t.Fatal(err)
	}
	if len(collection.Tags.Import) != 5 {
		t.Fatalf("starter imports = %#v, want five groups", collection.Tags.Import)
	}
	catalog, diagnostics := LoadCatalog(paths, app)
	if diagnostics.HasErrors() {
		t.Fatalf("starter catalog diagnostics = %#v", diagnostics)
	}
	if _, graphDiagnostics := BuildValidatedGraph(catalog); graphDiagnostics.HasErrors() {
		t.Fatalf("starter graph diagnostics = %#v", graphDiagnostics)
	}
	for _, group := range []string{"creator", "universe", "character", "general", "metadata"} {
		if _, ok := catalog.Group(group); !ok {
			t.Errorf("starter group %q is missing", group)
		}
	}
	character, _, ok := catalog.Tag("character")
	if !ok {
		t.Fatal("starter character tag is missing")
	}
	for _, alias := range []string{
		"konata",
		"izumi_konata",
		"泉こなた",
		"коната",
		"коната_изуми",
	} {
		canonical, found := CanonicalPredefinedValue(character, alias)
		if !found || canonical != "konata_izumi" {
			t.Errorf("character alias %q resolved to %q, %t", alias, canonical, found)
		}
	}
	konata := character.Values[1]
	if len(konata.Demand) != 1 || konata.Demand[0].Tag != "universe" ||
		len(konata.Demand[0].Has) != 1 || konata.Demand[0].Has[0] != "lucky_star" ||
		konata.Demand[0].Reason == "" {
		t.Fatalf("Konata demand = %#v", konata.Demand)
	}
}

func TestEnsureDefaultsUsesExistingAppDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `default_collection: pictures
default_storage_name: archive
default_storage_path: $HOME/archive
`
	if err := os.WriteFile(paths.App, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	app, err := EnsureDefaults(paths)
	if err != nil {
		t.Fatalf("EnsureDefaults() error = %v", err)
	}
	if app.DefaultCollection != "pictures" || app.DefaultStorageName != "archive" {
		t.Fatalf("EnsureDefaults() app = %#v", app)
	}
	if _, err := os.Stat(filepath.Join(paths.Collections, "pictures.yaml")); err != nil {
		t.Fatalf("expected custom default collection: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, "archive")); err != nil {
		t.Fatalf("expected custom default storage: %v", err)
	}
}

func TestEnsureDefaultsIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureStarterDefaults(paths); err != nil {
		t.Fatalf("first EnsureDefaults() error = %v", err)
	}
	before := readDefaultFiles(t, paths, "main")
	if _, err := EnsureStarterDefaults(paths); err != nil {
		t.Fatalf("second EnsureDefaults() error = %v", err)
	}
	after := readDefaultFiles(t, paths, "main")
	for path, want := range before {
		if got := after[path]; got != want {
			t.Errorf("EnsureDefaults() changed %q", path)
		}
	}
}

func TestStarterCharactersUseLuckyStarCast(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureStarterDefaults(paths); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(paths.Tags, "character.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, name := range []string{"konata_izumi", "kagami_hiiragi", "tsukasa_hiiragi", "miyuki_takara"} {
		if !strings.Contains(content, name) {
			t.Errorf("starter characters missing %q", name)
		}
	}
	if strings.Contains(content, "original_character") {
		t.Error("starter characters contain original_character")
	}
}

func TestEnsureDefaultsRejectsMissingDefaultStorage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := (YAMLFile[StorageConfig]{Path: paths.Storage}).Write(StorageConfig{{
		Name: "archive",
		Type: "local",
		Path: "$HOME/archive",
	}}); err != nil {
		t.Fatal(err)
	}
	_, err = EnsureDefaults(paths)
	if err == nil || !strings.Contains(err.Error(), "does not define default storage") {
		t.Fatalf("EnsureDefaults() error = %v, want missing default storage", err)
	}
}

func TestEnsureDefaultsRejectsConflictingCollection(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Collections, 0o755); err != nil {
		t.Fatal(err)
	}
	collection := DefaultCollectionConfig(DefaultAppConfig())
	collection.Location = filepath.Join(home, "another.sqlite")
	file := YAMLFile[CollectionConfig]{
		Path:     filepath.Join(paths.Collections, "main.yaml"),
		Validate: VerifyCollectionConfig,
	}
	if err := file.Write(collection); err != nil {
		t.Fatal(err)
	}
	_, err = EnsureDefaults(paths)
	if err == nil || !strings.Contains(err.Error(), "conflicts with freebooru.yaml") {
		t.Fatalf("EnsureDefaults() error = %v, want collection conflict", err)
	}
}

func TestEnsureDefaultsStopsAfterInvalidApp(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths, err := PathsFromDir(filepath.Join(home, "config"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(paths.App, []byte("http_port: 70000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := EnsureDefaults(paths); err == nil {
		t.Fatal("EnsureDefaults() succeeded with invalid app config")
	}
	for _, path := range []string{paths.Storage, paths.Tags, paths.Collections} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("unexpected domain resource %q after app failure", path)
		}
	}
}

func readDefaultFiles(t *testing.T, paths Paths, collection string) map[string]string {
	t.Helper()
	files := []string{
		paths.App,
		paths.Storage,
		filepath.Join(paths.Collections, collection+".yaml"),
		filepath.Join(paths.Tags, "character.yaml"),
		filepath.Join(paths.Tags, "creator.yaml"),
		filepath.Join(paths.Tags, "general.yaml"),
		filepath.Join(paths.Tags, "metadata.yaml"),
		filepath.Join(paths.Tags, "universe.yaml"),
	}
	contents := make(map[string]string, len(files))
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		contents[path] = string(data)
	}
	return contents
}
