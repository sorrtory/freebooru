package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func editAndCheck(cmd *cobra.Command, options *rootOptions, path string) error {
	editor := strings.Fields(os.Getenv("EDITOR"))
	if len(editor) == 0 {
		return fmt.Errorf("EDITOR is not set")
	}
	arguments := append(append([]string(nil), editor[1:]...), path)
	process := exec.CommandContext(cmd.Context(), editor[0], arguments...)
	process.Stdin = cmd.InOrStdin()
	process.Stdout = cmd.OutOrStdout()
	process.Stderr = cmd.ErrOrStderr()
	if err := process.Run(); err != nil {
		return fmt.Errorf("run editor for %q: %w", path, err)
	}
	if _, err := loadCore(cmd.Context(), options); err != nil {
		return fmt.Errorf("validate edited configuration: %w", err)
	}
	return nil
}
