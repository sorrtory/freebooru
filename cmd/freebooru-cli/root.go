package main

import (
	"fmt"
	"os"

	
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "freebooru-cli",
	Short: "freebooru-cli is a cli tool for interacting with the Freebooru API",
	Long:  "freebooru-cli is a cli tool for interacting with the Freebooru API - searching, downloading, and managing images.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Avoid initialization for commands that don't need the core.
		if cmd.Name() == "version" {
			return nil
		}

		app := createCore()

		cmd.SetContext(withCore(cmd.Context(), app))
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from Freebooru-CLI!")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing Freebooru-CLI '%s'\n", err)
		os.Exit(1)
	}
}
