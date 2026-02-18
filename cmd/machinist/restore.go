package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore environment from manifest",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "restore: not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}
