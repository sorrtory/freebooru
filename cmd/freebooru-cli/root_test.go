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
