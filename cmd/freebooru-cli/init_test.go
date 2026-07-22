package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)

	command := newRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetErr(new(bytes.Buffer))
	command.SetArgs([]string{"init"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := output.String(); got != "FreeBooru initialized\n" {
		t.Fatalf("Execute() output = %q", got)
	}
	configDir := filepath.Join(configHome, "freebooru")
	for _, path := range []string{
		filepath.Join(configDir, "freebooru.yaml"),
		filepath.Join(configDir, "storage.yaml"),
		filepath.Join(configDir, "collections", "main.yaml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %q: %v", path, err)
		}
	}
}

func TestConfigInitCommandDoesNotExist(t *testing.T) {
	command := newRootCommand()
	command.SetOut(new(bytes.Buffer))
	command.SetErr(new(bytes.Buffer))
	command.SetArgs([]string{"config", "init"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("Execute() error = %v, want unknown command", err)
	}
}

func TestInitCommandFailureDoesNotPrintSuccess(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", t.TempDir())
	configDir := filepath.Join(configHome, "freebooru")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(configDir, "freebooru.yaml"),
		[]byte("http_port: 70000\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	command := newRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetErr(new(bytes.Buffer))
	command.SetArgs([]string{"init"})
	if err := command.Execute(); err == nil {
		t.Fatal("Execute() succeeded with invalid application config")
	}
	if output.Len() != 0 {
		t.Fatalf("Execute() output = %q, want no success output", output.String())
	}
}
