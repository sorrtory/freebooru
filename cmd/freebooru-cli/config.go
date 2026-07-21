package main

import (
	"github.com/sorrtory/freebooru/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg"},
	Short:   "Work with configuration",
	Run: func(cmd *cobra.Command, args []string) {
		config.Load()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
