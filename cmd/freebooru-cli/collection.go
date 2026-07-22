package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Cobra cannot model a dynamic collection name before a subcommand directly.
// This command extracts the name and delegates the remaining arguments to a
// fresh collection-scoped command tree, preserving normal leaf validation.
func newCollectionCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:                "collection <name> <command>",
		Short:              "Run a command for an explicit collection",
		Args:               dynamicCommandArgs(2),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isHelpArgument(args[0]) {
				return cmd.Help()
			}
			name := args[0]
			if name == "" {
				return fmt.Errorf("collection name is required")
			}
			scope := newCollectionScopeCommand(options, name)
			return executeNestedCommand(cmd, scope, args[1:])
		},
	}
}

func dynamicCommandArgs(minimum int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && isHelpArgument(args[0]) {
			return nil
		}
		return cobra.MinimumNArgs(minimum)(cmd, args)
	}
}

func isHelpArgument(value string) bool {
	return value == "--help" || value == "-h"
}

func newCollectionScopeCommand(options *rootOptions, name string) *cobra.Command {
	command := &cobra.Command{
		Use:           name,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	command.AddCommand(
		newImportCommand(options, name),
		newTagCommand(options, name),
		newSearchCommand(options, name),
	)
	return command
}

func executeNestedCommand(parent, nested *cobra.Command, args []string) error {
	nested.SetArgs(args)
	nested.SetIn(parent.InOrStdin())
	nested.SetOut(parent.OutOrStdout())
	nested.SetErr(parent.ErrOrStderr())
	nested.SetContext(parent.Context())
	return nested.Execute()
}
