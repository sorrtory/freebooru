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
		Use:   "collection list | edit <name> | <name> <command>",
		Short: "List, edit, or explicitly select a collection",
		Long: "List or edit global collection definitions, or select a collection " +
			"for a nested command without changing default_collection.",
		Example: "  freebooru-cli collection list\n" +
			"  freebooru-cli collection edit archive\n" +
			"  freebooru-cli collection archive files list",
		Args:               collectionCommandArgs,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if isHelpArgument(args[0]) {
				return cmd.Help()
			}
			if args[0] == "list" || args[0] == "edit" {
				return executeNestedCommand(cmd, newCollectionCatalogCommand(options), args)
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

func collectionCommandArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 1 && (isHelpArgument(args[0]) || args[0] == "list") {
		return nil
	}
	return cobra.MinimumNArgs(2)(cmd, args)
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
		newStorageCommand(options, name, false),
		newFilesCommand(options, name),
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
