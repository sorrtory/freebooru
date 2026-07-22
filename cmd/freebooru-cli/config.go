// Package main provides the FreeBooru command-line application.
package main

import (
	"fmt"

	"github.com/sorrtory/freebooru/internal/bootstrap"
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
	command.AddCommand(
		newConfigInitCommand(options),
		newConfigCheckCommand(options),
	)
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
			if err := app.CheckConfig(cmd.Context()); err != nil {
				return fmt.Errorf("configuration check failed: %w", err)
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "Configuration is valid"); err != nil {
				return fmt.Errorf("write result: %w", err)
			}
			return nil
		},
	}
}

func newConfigInitCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create the default FreeBooru configuration layout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := bootstrap.NewCore(newLogger(options.verbose))
			if err != nil {
				return err
			}
			if err := app.InitConfig(cmd.Context()); err != nil {
				return fmt.Errorf("initialize configuration: %w", err)
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "FreeBooru configuration initialized"); err != nil {
				return fmt.Errorf("write result: %w", err)
			}
			return nil
		},
	}
}
