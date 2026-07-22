package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMVPCleanHomeWorkflow is the executable release workflow. Every command
// uses a fresh Cobra tree and Core, matching separate CLI invocations.
func TestMVPCleanHomeWorkflow(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	if output := executeCLI(t, []string{"init"}); output != "FreeBooru initialized\n" {
		t.Fatalf("init output = %q", output)
	}
	configDir := filepath.Join(configHome, "freebooru")
	if err := os.WriteFile(
		filepath.Join(configDir, "storage.yaml"),
		[]byte("- name: default\n  type: local\n  path: $HOME/.local/share/freebooru/storage/default\n- name: archive\n  type: local\n  path: $HOME/.local/share/freebooru/storage/archive\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(configDir, "tags", "reviewed.yaml"),
		[]byte("name: reviewed\ntype: bool\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(configDir, "collections", "main.yaml"),
		[]byte("name: main\ntags:\n  import:\n    - storage: default\n    - storage: archive\n    - tag: reviewed\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if output := executeCLI(t, []string{"config", "check"}); output != "Configuration is valid\n" {
		t.Fatalf("config check output = %q", output)
	}

	contents := []byte("MVP release workflow")
	source := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(source, contents, 0o600); err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(contents))
	if output := executeCLI(t, []string{"import", source, "--tag", "reviewed"}); output != hash+"\n" {
		t.Fatalf("import output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "get"}); output != "reviewed\nstorage:default\n" {
		t.Fatalf("tag get output = %q", output)
	}
	if output := executeCLI(t, []string{"search", "reviewed"}); output != hash+"\n" {
		t.Fatalf("search output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "remove", "reviewed"}); output != hash+"\n" {
		t.Fatalf("tag remove output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "add", "reviewed"}); output != hash+"\n" {
		t.Fatalf("tag add output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "add", "storage:archive"}); output != hash+"\n" {
		t.Fatalf("storage add output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "remove", "storage:default"}); output != hash+"\n" {
		t.Fatalf("default storage remove output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "get"}); output != "reviewed\nstorage:archive\n" {
		t.Fatalf("reopened tag get output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "remove", "storage:archive"}); output != hash+"\n" {
		t.Fatalf("final storage remove output = %q", output)
	}
	if output := executeCLI(t, []string{"search", "reviewed"}); output != "" {
		t.Fatalf("search after final removal output = %q", output)
	}
	for _, storage := range []string{"default", "archive"} {
		path := filepath.Join(home, ".local", "share", "freebooru", "storage", storage, hash[:2], hash)
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("storage %s copy remains: %v", storage, err)
		}
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatalf("source was removed with remove_on_upload false: %v", err)
	}

	command := newRootCommand()
	command.SetArgs([]string{"tag", hash, "get"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("tag get after final removal error = %v", err)
	}
}
