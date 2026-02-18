package main

import (
	"fmt"

	"github.com/moinsen-dev/machinist/profiles"
	"github.com/spf13/cobra"
)

var listProfilesCmd = &cobra.Command{
	Use:   "list-profiles",
	Short: "List available built-in profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		names, err := profiles.List()
		if err != nil {
			return err
		}
		for _, name := range names {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listProfilesCmd)
}
