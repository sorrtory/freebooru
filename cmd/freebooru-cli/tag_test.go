package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTagCommandsMutateAndReadTypedAssignments(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})

	configDir := filepath.Join(configHome, "freebooru")
	if err := os.WriteFile(
		filepath.Join(configDir, "tags", "rating.yaml"),
		[]byte("name: rating\ntype: value\nvalues:\n  - val: safe\n  - val: questionable\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(configDir, "collections", "main.yaml"),
		[]byte("name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - tag: rating\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	content := []byte("typed tag commands")
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(content))
	if output := executeCLI(t, []string{"import", path, "--tag", "rating:safe"}); output != hash+"\n" {
		t.Fatalf("import output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", strings.ToUpper(hash), "get"}); output != "rating:safe\nstorage:default\n" {
		t.Fatalf("tag get output = %q", output)
	}
	if output := executeCLI(t, []string{
		"collection", "main", "tag", hash, "set", "rating:questionable",
	}); output != hash+"\n" {
		t.Fatalf("tag set output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "get"}); output != "rating:questionable\nstorage:default\n" {
		t.Fatalf("tag get after set = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "remove", "rating"}); output != hash+"\n" {
		t.Fatalf("tag remove output = %q", output)
	}
	if output := executeCLI(t, []string{"tag", hash, "get"}); output != "storage:default\n" {
		t.Fatalf("tag get after remove = %q", output)
	}
}

func TestTagCommandRejectsInvalidHashWithoutUsageOutput(t *testing.T) {
	command := newRootCommand()
	output := new(bytes.Buffer)
	errors := new(bytes.Buffer)
	command.SetOut(output)
	command.SetErr(errors)
	command.SetArgs([]string{"tag", "not-a-hash", "get"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "exactly 64") {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("stdout = %q, stderr = %q", output.String(), errors.String())
	}
}

func TestTagCommandReturnsEvaluatorReasonWithoutUsageNoise(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})
	configDir := filepath.Join(configHome, "freebooru")
	if err := os.WriteFile(
		filepath.Join(configDir, "tags", "relationships.yaml"),
		[]byte("name: trigger\ntype: bool\ndemand:\n  - tag: title\n    reason: triggered files need a title\n---\nname: title\ntype: text\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(configDir, "collections", "main.yaml"),
		[]byte("name: main\ntags:\n  require:\n    - storage: default\n  import:\n    - tag: trigger\n    - tag: title\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	content := []byte("evaluator failure")
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(content))
	executeCLI(t, []string{"import", path})

	command := newRootCommand()
	output := new(bytes.Buffer)
	errors := new(bytes.Buffer)
	command.SetOut(output)
	command.SetErr(errors)
	command.SetArgs([]string{"tag", hash, "add", "trigger"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "triggered files need a title") {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.Len() != 0 || errors.Len() != 0 {
		t.Fatalf("stdout = %q, stderr = %q", output.String(), errors.String())
	}
}
