package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestAssignmentCompletionUsesCheckedCollectionFieldsAndPersistedState(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})
	configDir := filepath.Join(configHome, "freebooru")
	if err := os.Remove(filepath.Join(configDir, "tags", "general.yaml")); err != nil {
		t.Fatal(err)
	}
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

	options := &rootOptions{newCore: defaultCoreFactory}
	importCommand := newImportCommand(options, "")
	importCommand.SetContext(t.Context())
	completeFlag, ok := importCommand.GetFlagCompletionFunc("tag")
	if !ok {
		t.Fatal("--tag completion is not registered")
	}
	suggestions, directive := completeFlag(importCommand, nil, "rat")
	if !reflect.DeepEqual(suggestions, []string{"rating:"}) ||
		directive&cobra.ShellCompDirectiveNoFileComp == 0 {
		t.Fatalf("tag name completion = %#v, %v", suggestions, directive)
	}
	suggestions, _ = completeFlag(importCommand, nil, "rating:q")
	if !reflect.DeepEqual(suggestions, []string{"rating:questionable"}) {
		t.Fatalf("tag value completion = %#v", suggestions)
	}

	content := []byte("completion state")
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(content))
	executeCLI(t, []string{"import", path, "--tag", "rating:safe"})
	mutation := newTagAssignmentCommand(options, "", hash, "set")
	mutation.SetContext(t.Context())
	suggestions, directive = mutation.ValidArgsFunction(mutation, nil, "rating:q")
	if !reflect.DeepEqual(suggestions, []string{"rating:questionable"}) ||
		directive&cobra.ShellCompDirectiveNoFileComp == 0 {
		t.Fatalf("persisted value completion = %#v, %v", suggestions, directive)
	}
}
