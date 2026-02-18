package main

import (
	"fmt"

	"github.com/moinsen-dev/machinist/profiles"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scanners and profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := newRegistry()

		fmt.Fprintln(cmd.OutOrStdout(), "Scanners:")
		for _, s := range reg.List() {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", s.Name(), s.Description())
		}

		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Profiles:")
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
	rootCmd.AddCommand(listCmd)
}
