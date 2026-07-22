package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFilesCommand(options *rootOptions, collectionName string) *cobra.Command {
	command := &cobra.Command{Use: "files", Short: "Inspect indexed collection files", Args: cobra.NoArgs}
	var storage string
	list := &cobra.Command{
		Use:   "list",
		Short: "List one SHA-256 and storage pair per physical copy",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := loadCore(cmd.Context(), options)
			if err != nil {
				return err
			}
			rows, err := app.ListFiles(cmd.Context(), collectionName, storage)
			if err != nil {
				return err
			}
			for _, row := range rows {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", row.SHA256, row.Storage); err != nil {
					return fmt.Errorf("write file storage row: %w", err)
				}
			}
			return nil
		},
	}
	list.Flags().StringVar(&storage, "storage", "", "only files assigned to this imported storage")
	command.AddCommand(list)
	return command
}
