package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCollectionCatalogCommand(options *rootOptions) *cobra.Command {
	command := &cobra.Command{Use: "collection", SilenceUsage: true, SilenceErrors: true}
	command.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List configured collections",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := loadCore(cmd.Context(), options)
				if err != nil {
					return err
				}
				collections, err := app.ListCollections()
				if err != nil {
					return err
				}
				for _, item := range collections {
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), item.Name); err != nil {
						return fmt.Errorf("write collection name: %w", err)
					}
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "edit <name>",
			Short: "Edit the YAML file defining a collection",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				app, err := loadCore(cmd.Context(), options)
				if err != nil {
					return err
				}
				path, err := app.CollectionConfigSource(args[0])
				if err != nil {
					return err
				}
				return editAndCheck(cmd, options, path)
			},
		},
	)
	return command
}

func newTagCatalogCommand(options *rootOptions, collectionName string) *cobra.Command {
	command := &cobra.Command{Use: "tag", SilenceUsage: true, SilenceErrors: true}
	command.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List tags imported by the selected collection",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				app, err := loadCore(cmd.Context(), options)
				if err != nil {
					return err
				}
				tags, err := app.ListCollectionTags(collectionName)
				if err != nil {
					return err
				}
				for _, tag := range tags {
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), tag.Name); err != nil {
						return fmt.Errorf("write tag name: %w", err)
					}
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "edit <name>",
			Short: "Edit the YAML file defining an imported tag",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				app, err := loadCore(cmd.Context(), options)
				if err != nil {
					return err
				}
				path, err := app.TagConfigSource(collectionName, args[0])
				if err != nil {
					return err
				}
				return editAndCheck(cmd, options, path)
			},
		},
	)
	return command
}
