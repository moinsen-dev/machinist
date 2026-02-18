package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/spf13/cobra"
)

var (
	snapshotOutput      string
	snapshotInteractive bool
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Scan environment and generate TOML manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := newRegistry()
		ctx := context.Background()

		if snapshotInteractive {
			snap, errs, err := runInteractiveScan(cmd, reg, ctx)
			if err != nil {
				return err
			}
			if snap == nil {
				return nil // cancelled
			}
			for _, e := range errs {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", e)
			}
			if err := domain.WriteManifest(snap, snapshotOutput); err != nil {
				return fmt.Errorf("write manifest: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Snapshot written to %s\n", snapshotOutput)
			return nil
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Scanning environment...")
		progress := newProgressWriter(cmd.OutOrStdout())
		snap, errs := reg.ScanAllWithProgress(ctx, progress)
		for _, e := range errs {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", e)
		}
		fmt.Fprintln(cmd.OutOrStdout())

		if err := domain.WriteManifest(snap, snapshotOutput); err != nil {
			return fmt.Errorf("write manifest: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Snapshot written to %s (%.1fs)\n", snapshotOutput, snap.Meta.ScanDurationSecs)
		return nil
	},
}

// countSections counts the number of non-nil section pointers in a Snapshot.
func countSections(snap *domain.Snapshot) int {
	v := reflect.ValueOf(snap).Elem()
	count := 0
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		name := v.Type().Field(i).Name
		if name == "Meta" {
			continue
		}
		if f.Kind() == reflect.Ptr && !f.IsNil() {
			count++
		}
	}
	return count
}

// sectionNames returns the TOML key names of non-nil sections in a Snapshot.
func sectionNames(snap *domain.Snapshot) []string {
	v := reflect.ValueOf(snap).Elem()
	t := v.Type()
	var names []string
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		sf := t.Field(i)
		if sf.Name == "Meta" {
			continue
		}
		if f.Kind() == reflect.Ptr && !f.IsNil() {
			tag := sf.Tag.Get("toml")
			name := strings.SplitN(tag, ",", 2)[0]
			if name == "" {
				name = sf.Name
			}
			names = append(names, name)
		}
	}
	return names
}

func init() {
	snapshotCmd.Flags().StringVarP(&snapshotOutput, "output", "o", "machinist-snapshot.toml", "Output file path")
	snapshotCmd.Flags().BoolVarP(&snapshotInteractive, "interactive", "i", false, "Interactively select which scanners to run")
	rootCmd.AddCommand(snapshotCmd)
}
