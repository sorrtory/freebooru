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
		Args:               cobra.MinimumNArgs(2),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("collection name is required")
			}
			scope := newCollectionScopeCommand(options, name)
			return executeNestedCommand(cmd, scope, args[1:])
		},
	}
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
