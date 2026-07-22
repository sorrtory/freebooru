package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSearchCommandsReturnPersistedMatches(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})

	configDir := filepath.Join(configHome, "freebooru")
	if err := os.WriteFile(
		filepath.Join(configDir, "tags", "reviewed.yaml"),
		[]byte("name: reviewed\ntype: bool\n---\nname: category\ntype: value\nvalues:\n  - val: safe\n    aliases: [sfw]\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(configDir, "collections", "main.yaml"),
		[]byte("name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - tag: reviewed\n    - tag: category\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	matched := []byte("reviewed search result")
	matchedPath := writeSearchFixture(t, matched)
	matchedHash := fmt.Sprintf("%x", sha256.Sum256(matched))
	executeCLI(t, []string{"import", matchedPath, "--tag", "reviewed", "--tag", "category:sfw"})
	unmatchedPath := writeSearchFixture(t, []byte("not reviewed"))
	executeCLI(t, []string{"import", unmatchedPath})

	if output := executeCLI(t, []string{"search", "reviewed"}); output != matchedHash+"\n" {
		t.Fatalf("default search output = %q", output)
	}
	if output := executeCLI(t, []string{"search", "category:SFW"}); output != matchedHash+"\n" {
		t.Fatalf("alias search output = %q", output)
	}
	if output := executeCLI(t, []string{"search", "sha256:" + matchedHash}); output != matchedHash+"\n" {
		t.Fatalf("SHA-256 system search output = %q", output)
	}
	if output := executeCLI(t, []string{"search", fmt.Sprintf("filesize:%d", len(matched))}); output != matchedHash+"\n" {
		t.Fatalf("filesize system search output = %q", output)
	}
	if output := executeCLI(t, []string{"search", "filetype:text/plain; charset=utf-8", "reviewed"}); output != matchedHash+"\n" {
		t.Fatalf("filetype system search output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", matchedHash, "get"}); !strings.Contains(output, "category:safe\n") {
		t.Fatalf("canonical tag output = %q", output)
	}
	if output := executeCLI(t, []string{
		"collection", "main", "search", "reviewed", "--limit", "1", "--offset", "0",
	}); output != matchedHash+"\n" {
		t.Fatalf("explicit search output = %q", output)
	}
}

func TestSearchCommandRejectsNegativePagination(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})
	command := newRootCommand()
	command.SetArgs([]string{"search", "storage:default", "--limit", "-1"})
	if err := command.Execute(); err == nil {
		t.Fatal("negative search limit succeeded")
	}
}

func writeSearchFixture(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
