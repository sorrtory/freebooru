package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportCommandsUseDefaultAndExplicitCollections(t *testing.T) {
	for _, test := range []struct {
		name string
		args func(string) []string
	}{
		{name: "default", args: func(path string) []string { return []string{"import", path} }},
		{
			name: "explicit",
			args: func(path string) []string { return []string{"collection", "main", "import", path} },
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			configHome := t.TempDir()
			home := t.TempDir()
			t.Setenv("XDG_CONFIG_HOME", configHome)
			t.Setenv("HOME", home)
			executeCLI(t, []string{"init"})

			content := []byte("freebooru " + test.name)
			path := filepath.Join(t.TempDir(), "source.bin")
			if err := os.WriteFile(path, content, 0o600); err != nil {
				t.Fatal(err)
			}
			output := executeCLI(t, test.args(path))
			want := fmt.Sprintf("%x\n", sha256.Sum256(content))
			if output != want {
				t.Fatalf("output = %q, want %q", output, want)
			}
		})
	}
}

func TestImportHelpExplainsAssignmentsAndCollectionSelection(t *testing.T) {
	command := newRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"import", "--help"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, text := range []string{
		"Boolean tags use --tag name",
		"--tag labels:portrait --tag labels:outdoors",
		"--interactive",
		"collection archive import",
	} {
		if !strings.Contains(output.String(), text) {
			t.Fatalf("import help does not contain %q: %s", text, output.String())
		}
	}
}

func TestImportCommandValidatesArgumentsAndAssignments(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})

	command := newRootCommand()
	command.SetOut(new(bytes.Buffer))
	command.SetErr(new(bytes.Buffer))
	command.SetArgs([]string{"import"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "arg(s)") {
		t.Fatalf("missing path error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, []byte("invalid assignment"), 0o600); err != nil {
		t.Fatal(err)
	}
	command = newRootCommand()
	command.SetOut(new(bytes.Buffer))
	command.SetErr(new(bytes.Buffer))
	command.SetArgs([]string{"import", path, "--tag", "unknown:value"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "not imported") {
		t.Fatalf("invalid assignment error = %v", err)
	}
}

func TestInteractiveImportPromptsForRequiredThenOptionalTags(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})
	configDir := filepath.Join(configHome, "freebooru")
	if err := os.WriteFile(
		filepath.Join(configDir, "tags", "fields.yaml"),
		[]byte("name: rating\ntype: value\nvalues:\n  - val: safe\n---\nname: reviewed\ntype: bool\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(configDir, "collections", "main.yaml"),
		[]byte("name: main\ntags:\n  require:\n    - storage: default\n    - tag: rating\n  import:\n    - tag: reviewed\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	content := []byte("interactive import")
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	command := newRootCommand()
	output := new(bytes.Buffer)
	prompts := new(bytes.Buffer)
	command.SetIn(strings.NewReader("safe\nyes\n"))
	command.SetOut(output)
	command.SetErr(prompts)
	command.SetArgs([]string{"import", path, "--interactive"})
	if err := command.Execute(); err != nil {
		t.Fatalf("interactive import error = %v", err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(content))
	if output.String() != hash+"\n" {
		t.Fatalf("interactive output = %q", output.String())
	}
	if got := prompts.String(); !strings.Contains(got, "required rating") ||
		!strings.Contains(got, "optional reviewed") ||
		strings.Index(got, "required rating") > strings.Index(got, "optional reviewed") {
		t.Fatalf("prompts = %q", got)
	}
	if got := executeCLI(t, []string{"tag", hash, "get"}); got != "rating:safe\nreviewed\nstorage:default\n" {
		t.Fatalf("persisted assignments = %q", got)
	}
}

func TestImportCommandPropagatesCancellation(t *testing.T) {
	configHome := t.TempDir()
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("HOME", home)
	executeCLI(t, []string{"init"})
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, []byte("canceled import"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	command := newRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetErr(new(bytes.Buffer))
	command.SetArgs([]string{"import", path})
	command.SetContext(ctx)
	if err := command.Execute(); !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() error = %v, want context.Canceled", err)
	}
	if output.Len() != 0 {
		t.Fatalf("canceled output = %q", output.String())
	}
}

func executeCLI(t *testing.T, args []string) string {
	t.Helper()
	command := newRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetErr(new(bytes.Buffer))
	command.SetArgs(args)
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute(%q) error = %v", args, err)
	}
	return output.String()
}
