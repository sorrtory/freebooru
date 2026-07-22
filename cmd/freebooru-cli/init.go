package main

import (
	"fmt"

	"github.com/sorrtory/freebooru/internal/bootstrap"
	"github.com/spf13/cobra"
)

func newInitCommand(options *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize FreeBooru",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := bootstrap.NewCore(newLogger(options.verbose))
			if err != nil {
				return err
			}
			if err := app.Init(cmd.Context()); err != nil {
				return fmt.Errorf("initialize FreeBooru: %w", err)
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "FreeBooru initialized"); err != nil {
				return fmt.Errorf("write result: %w", err)
			}
			return nil
		},
	}
}
