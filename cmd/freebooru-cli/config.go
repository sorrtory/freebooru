// Package main provides the FreeBooru command-line application.
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sorrtory/freebooru/internal/bootstrap"
	appconfig "github.com/sorrtory/freebooru/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCommand(options *rootOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "Work with configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	command.AddCommand(newConfigCheckCommand(options))
	return command
}

func newConfigCheckCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check storage, tag, and collection configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := bootstrap.NewCore(newLogger(options.verbose))
			if err != nil {
				return err
			}
			if err := app.LoadConfig(cmd.Context()); err != nil {
				return err
			}
			diagnostics := app.CheckConfig(cmd.Context())
			if diagnostics.HasErrors() {
				return configDiagnosticsError{diagnostics: diagnostics}
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid"); err != nil {
				return fmt.Errorf("write result: %w", err)
			}
			return nil
		},
	}
}

type configDiagnosticsError struct {
	diagnostics appconfig.Diagnostics
}

func (e configDiagnosticsError) Error() string {
	var message strings.Builder
	message.WriteString("configuration check failed")
	for _, diagnostic := range e.diagnostics {
		message.WriteString("\n")
		message.WriteString(diagnostic.File)
		if diagnostic.Document > 0 {
			message.WriteString(": document ")
			message.WriteString(strconv.Itoa(diagnostic.Document))
		}
		if diagnostic.Field != "" {
			message.WriteString(": ")
			message.WriteString(diagnostic.Field)
		}
		message.WriteString(": [")
		message.WriteString(diagnostic.Code)
		message.WriteString("] ")
		message.WriteString(diagnostic.Message)
	}
	return message.String()
}
