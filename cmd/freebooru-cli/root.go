package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

type rootOptions struct {
	verbose    bool
	newCore    coreFactory
	runCommand commandRunner
}

type commandRunner func(context.Context, io.Reader, io.Writer, io.Writer, string, ...string) error

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	options := &rootOptions{newCore: defaultCoreFactory, runCommand: runExternalCommand}
	return newRootCommandWithOptions(options)
}

func newRootCommandWithOptions(options *rootOptions) *cobra.Command {
	root := &cobra.Command{
		Use:           "freebooru",
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
		newSearchCommand(options, ""),
		newCollectionCommand(options),
		newStorageCommand(options, "", true),
		newFilesCommand(options, ""),
		newServerCommand(options),
	)

	return root
}

func runExternalCommand(
	ctx context.Context,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	name string,
	args ...string,
) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = stderr
	return command.Run()
}

func newLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
}
