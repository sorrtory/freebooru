package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogAndFileListCommands(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})
	configDir := filepath.Join(configHome, "freebooru")
	writeCommandFixture(t, filepath.Join(configDir, "tags", "reviewed.yaml"), "name: reviewed\ntype: bool\n")
	writeCommandFixture(
		t,
		filepath.Join(configDir, "collections", "main.yaml"),
		"name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - tag: reviewed\n",
	)

	if output := executeCLI(t, []string{"collection", "list"}); output != "main\n" {
		t.Fatalf("collection list output = %q", output)
	}
	if output := executeCLI(t, []string{"storage", "list"}); output != "default\n" {
		t.Fatalf("storage list output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", "list"}); output != "reviewed\nstorage\n" {
		t.Fatalf("tag list output = %q", output)
	}
	if output := executeCLI(t, []string{"collection", "main", "tag", "list"}); output != "reviewed\nstorage\n" {
		t.Fatalf("explicit tag list output = %q", output)
	}

	content := []byte("listed file")
	path := filepath.Join(t.TempDir(), "listed.bin")
	writeCommandFixture(t, path, string(content))
	hash := fmt.Sprintf("%x", sha256.Sum256(content))
	executeCLI(t, []string{"import", path})
	want := hash + "\tdefault\n"
	if output := executeCLI(t, []string{"files", "list"}); output != want {
		t.Fatalf("files list output = %q", output)
	}
	if output := executeCLI(t, []string{"collection", "main", "files", "list", "--storage", "default"}); output != want {
		t.Fatalf("filtered explicit files list output = %q", output)
	}
	command := newRootCommand()
	command.SetArgs([]string{"files", "list", "--storage", "missing"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "not imported") {
		t.Fatalf("missing storage error = %v", err)
	}
}

func TestEditCommandsOpenExpectedSources(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})
	configDir := filepath.Join(configHome, "freebooru")
	tagPath := filepath.Join(configDir, "tags", "reviewed.yaml")
	writeCommandFixture(t, tagPath, "name: reviewed\ntype: bool\n")
	collectionPath := filepath.Join(configDir, "collections", "main.yaml")
	writeCommandFixture(t, collectionPath, "name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - tag: reviewed\n")

	marker := filepath.Join(t.TempDir(), "edited-path")
	editor := filepath.Join(t.TempDir(), "editor")
	writeCommandFixture(t, editor, "#!/bin/sh\nprintf '%s' \"$1\" > \"$EDITOR_MARKER\"\n")
	if err := os.Chmod(editor, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", editor)
	t.Setenv("EDITOR_MARKER", marker)

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"storage", "edit"}, filepath.Join(configDir, "storage.yaml")},
		{[]string{"collection", "edit", "main"}, collectionPath},
		{[]string{"tag", "edit", "reviewed"}, tagPath},
		{[]string{"collection", "main", "tag", "edit", "reviewed"}, tagPath},
	}
	for _, test := range tests {
		executeCLI(t, test.args)
		got, err := os.ReadFile(marker)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != test.want {
			t.Fatalf("%v opened %q, want %q", test.args, got, test.want)
		}
	}
}

func TestEditCommandRejectsInvalidResult(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})
	editor := filepath.Join(t.TempDir(), "editor")
	writeCommandFixture(t, editor, "#!/bin/sh\nprintf 'invalid: true\\n' > \"$1\"\n")
	if err := os.Chmod(editor, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EDITOR", editor)
	command := newRootCommand()
	command.SetArgs([]string{"storage", "edit"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "validate edited configuration") {
		t.Fatalf("invalid edit error = %v", err)
	}
}

func writeCommandFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
