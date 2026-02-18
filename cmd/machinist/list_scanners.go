package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listScannersCmd = &cobra.Command{
	Use:   "list-scanners",
	Short: "List available scanners",
	Run: func(cmd *cobra.Command, args []string) {
		reg := newRegistry()
		for _, s := range reg.List() {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", s.Name(), s.Description())
		}
	},
}

func init() {
	rootCmd.AddCommand(listScannersCmd)
}
