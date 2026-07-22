package main

import (
	"fmt"

	"github.com/sorrtory/freebooru/internal/core"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	limit  int64
	offset int64
}

func newSearchCommand(root *rootOptions, collectionName string) *cobra.Command {
	options := &searchOptions{limit: core.DefaultSearchLimit}
	command := &cobra.Command{
		Use:   "search <term>...",
		Short: "Search files using typed AND terms",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := loadCore(cmd.Context(), root)
			if err != nil {
				return err
			}
			files, err := app.Search(cmd.Context(), core.FileSearchRequest{
				Collection: collectionName,
				Terms:      args,
				Limit:      &options.limit,
				Offset:     options.offset,
			})
			if err != nil {
				return err
			}
			for _, file := range files {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), file.SHA256); err != nil {
					return fmt.Errorf("write search result: %w", err)
				}
			}
			return nil
		},
	}
	command.Flags().Int64Var(&options.limit, "limit", core.DefaultSearchLimit, "maximum result count")
	command.Flags().Int64Var(&options.offset, "offset", 0, "result offset")
	return command
}
