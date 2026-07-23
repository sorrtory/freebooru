package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const serverService = "freebooru-server.service"

func newServerCommand(options *rootOptions) *cobra.Command {
	command := &cobra.Command{
		Use:   "server",
		Short: "Control the packaged per-user HTTP server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	for _, action := range []string{"start", "stop", "restart", "status"} {
		action := action
		command.AddCommand(&cobra.Command{
			Use:   action,
			Short: action + " the per-user HTTP server",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return runServiceAction(cmd, options, action)
			},
		})
	}
	var follow bool
	logs := &cobra.Command{
		Use:   "logs",
		Short: "Show logs from the per-user HTTP server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			args := []string{"--user", "--unit", serverService}
			if follow {
				args = append(args, "--follow")
			}
			return runServiceCommand(cmd, options, "journalctl", args...)
		},
	}
	logs.Flags().BoolVarP(&follow, "follow", "f", false, "follow new log entries")
	command.AddCommand(logs)
	return command
}

func runServiceAction(cmd *cobra.Command, options *rootOptions, action string) error {
	switch action {
	case "start":
		if err := runServiceCommand(cmd, options, "systemctl", "--user", "daemon-reload"); err != nil {
			return err
		}
		return runServiceCommand(cmd, options, "systemctl", "--user", "enable", "--now", serverService)
	case "stop":
		return runServiceCommand(cmd, options, "systemctl", "--user", "disable", "--now", serverService)
	default:
		return runServiceCommand(cmd, options, "systemctl", "--user", action, serverService)
	}
}

func runServiceCommand(
	cmd *cobra.Command,
	options *rootOptions,
	name string,
	args ...string,
) error {
	if err := options.runCommand(
		cmd.Context(),
		cmd.InOrStdin(),
		cmd.OutOrStdout(),
		cmd.ErrOrStderr(),
		name,
		args...,
	); err != nil {
		return fmt.Errorf("run %s: %w", name, err)
	}
	return nil
}
