package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStorageCommand(options *rootOptions, collectionName string, allowEdit bool) *cobra.Command {
	command := &cobra.Command{
		Use:   "storage",
		Short: "List storage providers available to a collection",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List imported storage providers",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := loadCore(cmd.Context(), options)
			if err != nil {
				return err
			}
			storages, err := app.ListCollectionStorages(collectionName)
			if err != nil {
				return err
			}
			for _, storage := range storages {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), storage.Name); err != nil {
					return fmt.Errorf("write storage name: %w", err)
				}
			}
			return nil
		},
	})
	if allowEdit {
		command.AddCommand(&cobra.Command{
			Use:   "edit",
			Short: "Edit the global storage.yaml file",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := loadCore(cmd.Context(), options)
				if err != nil {
					return err
				}
				path, err := app.StorageConfigSource()
				if err != nil {
					return err
				}
				return editAndCheck(cmd, options, path)
			},
		})
	}
	return command
}
