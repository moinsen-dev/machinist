package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	composeFrom   string
	composeOutput string
)

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Compose a manifest from multiple sources",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "compose: not yet implemented")
	},
}

func init() {
	composeCmd.Flags().StringVar(&composeFrom, "from", "", "Source manifest file")
	composeCmd.Flags().StringVarP(&composeOutput, "output", "o", "composed-manifest.toml", "Output file path")
	rootCmd.AddCommand(composeCmd)
}
