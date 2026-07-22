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
