package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/moinsen-dev/machinist/internal/bundler"
	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/spf13/cobra"
)

var (
	snapshotOutput     string
	snapshotDryRun     bool
	snapshotFormat     string
	snapshotPassword   string
	snapshotVolumeName string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Scan environment and generate manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		if snapshotFormat != "toml" && snapshotFormat != "dmg" {
			return fmt.Errorf("unsupported format %q: must be \"toml\" or \"dmg\"", snapshotFormat)
		}

		reg := newRegistry()
		ctx := context.Background()

		if snapshotDryRun {
			scanners := reg.List()
			fmt.Fprintf(cmd.OutOrStdout(), "Dry-run mode: scanning environment\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Registered scanners: %d\n", len(scanners))
			for _, s := range scanners {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s (%s)\n", s.Name(), s.Category())
			}
			fmt.Fprintln(cmd.OutOrStdout())

			snap, errs := reg.ScanAll(ctx)
			for _, e := range errs {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", e)
			}

			sections := countSections(snap)
			fmt.Fprintf(cmd.OutOrStdout(), "Sections found: %d\n", sections)
			fmt.Fprintf(cmd.OutOrStdout(), "Output format: %s\n", snapshotFormat)
			fmt.Fprintf(cmd.OutOrStdout(), "Estimated restore stages: %d\n", sections)
			fmt.Fprintln(cmd.OutOrStdout())

			data, err := domain.MarshalManifest(snap)
			if err != nil {
				return fmt.Errorf("marshal: %w", err)
			}
			cmd.Print(string(data))
			return nil
		}

		snap, errs := reg.ScanAll(ctx)
		for _, e := range errs {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", e)
		}

		switch snapshotFormat {
		case "dmg":
			runner := &util.RealCommandRunner{}
			opts := bundler.BundleOptions{
				Password:   snapshotPassword,
				VolumeName: snapshotVolumeName,
			}
			if err := bundler.Bundle(ctx, runner, snap, snapshotOutput, opts); err != nil {
				return fmt.Errorf("create DMG bundle: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "DMG bundle written to %s\n", snapshotOutput)
		default:
			if err := domain.WriteManifest(snap, snapshotOutput); err != nil {
				return fmt.Errorf("write manifest: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Snapshot written to %s\n", snapshotOutput)
		}
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
	snapshotCmd.Flags().BoolVar(&snapshotDryRun, "dry-run", false, "Print TOML to stdout instead of writing file")
	snapshotCmd.Flags().StringVar(&snapshotFormat, "format", "toml", "Output format: \"toml\" or \"dmg\"")
	snapshotCmd.Flags().StringVar(&snapshotPassword, "password", "", "Optional DMG encryption password (only with --format=dmg)")
	snapshotCmd.Flags().StringVar(&snapshotVolumeName, "volume-name", "machinist", "DMG volume name (only with --format=dmg)")
	rootCmd.AddCommand(snapshotCmd)
}
