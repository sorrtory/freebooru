package main

import (
	"fmt"

	"github.com/sorrtory/freebooru/internal/core"
	"github.com/spf13/cobra"
)

type importOptions struct {
	tags        []string
	interactive bool
}

func newImportCommand(root *rootOptions, collectionName string) *cobra.Command {
	options := &importOptions{}
	command := &cobra.Command{
		Use:   "import <file>",
		Short: "Import one regular file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.interactive {
				return fmt.Errorf("interactive import is not implemented yet")
			}
			app, err := loadCore(cmd.Context(), root)
			if err != nil {
				return err
			}
			assignments, err := app.ParseTagAssignments(collectionName, options.tags)
			if err != nil {
				return err
			}
			result, err := app.Import(cmd.Context(), core.ImportRequest{
				Collection: collectionName,
				SourcePath: args[0],
				Tags:       assignments,
			})
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), result.SHA256); err != nil {
				return fmt.Errorf("write imported SHA-256: %w", err)
			}
			return nil
		},
	}
	command.Flags().StringArrayVar(
		&options.tags,
		"tag",
		nil,
		"assign a tag as name[:value] (repeatable)",
	)
	command.Flags().BoolVar(
		&options.interactive,
		"interactive",
		false,
		"prompt for missing and optional tags",
	)
	return command
}
