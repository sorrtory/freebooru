package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDynamicCommandsPrintHelpWithoutConfiguration(t *testing.T) {
	for _, name := range []string{"collection", "tag"} {
		t.Run(name, func(t *testing.T) {
			command := newRootCommand()
			output := new(bytes.Buffer)
			command.SetOut(output)
			command.SetErr(new(bytes.Buffer))
			command.SetArgs([]string{name, "--help"})
			if err := command.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if !strings.Contains(output.String(), "Usage:") {
				t.Fatalf("help output = %q", output.String())
			}
		})
	}
}

func TestSearchHelpExplainsTypedANDTerms(t *testing.T) {
	command := newRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"search", "--help"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	for _, text := range []string{"Every term must match", "!blocked", "score>=10"} {
		if !strings.Contains(output.String(), text) {
			t.Fatalf("search help does not contain %q: %s", text, output.String())
		}
	}
}
