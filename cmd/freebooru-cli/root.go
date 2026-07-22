package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

type rootOptions struct {
	verbose bool
	newCore coreFactory
}

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	options := &rootOptions{newCore: defaultCoreFactory}
	root := &cobra.Command{
		Use:           "freebooru-cli",
		Short:         "Interact with FreeBooru",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.PersistentFlags().BoolVarP(
		&options.verbose,
		"verbose",
		"v",
		false,
		"enable debug logging",
	)
	root.AddCommand(
		newInitCommand(options),
		newConfigCommand(options),
		newImportCommand(options, ""),
		newTagCommand(options, ""),
		newCollectionCommand(options),
	)

	return root
}

func newLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}
