package main

import (
	"context"
	"fmt"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [scanner-name]",
	Short: "Run a single scanner",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := newRegistry()
		ctx := context.Background()

		result, err := reg.ScanOne(ctx, args[0])
		if err != nil {
			return err
		}

		snap := domain.NewSnapshot("", "", "", Version)
		applyResultToSnapshot(snap, result)

		data, err := domain.MarshalManifest(snap)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		cmd.Print(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
