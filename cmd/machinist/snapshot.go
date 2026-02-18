package main

import (
	"context"
	"fmt"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/spf13/cobra"
)

var (
	snapshotOutput string
	snapshotDryRun bool
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Scan environment and generate manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := newRegistry()
		ctx := context.Background()

		snap, errs := reg.ScanAll(ctx)
		for _, e := range errs {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", e)
		}

		if snapshotDryRun {
			data, err := domain.MarshalManifest(snap)
			if err != nil {
				return fmt.Errorf("marshal: %w", err)
			}
			cmd.Print(string(data))
			return nil
		}

		if err := domain.WriteManifest(snap, snapshotOutput); err != nil {
			return fmt.Errorf("write manifest: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Snapshot written to %s\n", snapshotOutput)
		return nil
	},
}

func init() {
	snapshotCmd.Flags().StringVarP(&snapshotOutput, "output", "o", "machinist-snapshot.toml", "Output file path")
	snapshotCmd.Flags().BoolVar(&snapshotDryRun, "dry-run", false, "Print TOML to stdout instead of writing file")
	rootCmd.AddCommand(snapshotCmd)
}
