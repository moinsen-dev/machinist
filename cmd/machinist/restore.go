package main

import (
	"fmt"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/spf13/cobra"
)

var (
	restoreFrom   string
	restoreSkip   string
	restoreOnly   string
	restoreDryRun bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore environment from manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		if restoreFrom == "" {
			return fmt.Errorf("--from flag is required: provide a path to manifest.toml or mounted DMG")
		}

		if restoreSkip != "" && restoreOnly != "" {
			return fmt.Errorf("--skip and --only are mutually exclusive; use one or the other")
		}

		snap, err := domain.ReadManifest(restoreFrom)
		if err != nil {
			return fmt.Errorf("read manifest: %w", err)
		}

		sections := sectionNames(snap)

		// Apply --only filter
		if restoreOnly != "" {
			only := parseCSV(restoreOnly)
			sections = filterInclude(sections, only)
		}

		// Apply --skip filter
		if restoreSkip != "" {
			skip := parseCSV(restoreSkip)
			sections = filterExclude(sections, skip)
		}

		if restoreDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "Dry-run mode: restore plan\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Manifest: %s\n", restoreFrom)
			fmt.Fprintf(cmd.OutOrStdout(), "Host: %s (%s)\n", snap.Meta.SourceHostname, snap.Meta.SourceArch)
			fmt.Fprintf(cmd.OutOrStdout(), "Stages to execute: %d\n", len(sections))
			for i, s := range sections {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s\n", i+1, s)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nNo changes were made (dry-run).")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Restoring %d stages from %s\n", len(sections), restoreFrom)
		for _, s := range sections {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", s)
		}
		// TODO: execute the generated install.command / restore script for each stage
		fmt.Fprintln(cmd.OutOrStdout(), "\nRestore execution is not yet implemented.")
		return nil
	},
}

// parseCSV splits a comma-separated string into trimmed, non-empty tokens.
func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// filterInclude returns only the items present in the include list.
func filterInclude(items, include []string) []string {
	set := make(map[string]bool, len(include))
	for _, v := range include {
		set[v] = true
	}
	var result []string
	for _, item := range items {
		if set[item] {
			result = append(result, item)
		}
	}
	return result
}

// filterExclude returns items not present in the exclude list.
func filterExclude(items, exclude []string) []string {
	set := make(map[string]bool, len(exclude))
	for _, v := range exclude {
		set[v] = true
	}
	var result []string
	for _, item := range items {
		if !set[item] {
			result = append(result, item)
		}
	}
	return result
}

func init() {
	restoreCmd.Flags().StringVar(&restoreFrom, "from", "", "Path to manifest.toml or mounted DMG directory")
	restoreCmd.Flags().StringVar(&restoreSkip, "skip", "", "Comma-separated list of stages to skip")
	restoreCmd.Flags().StringVar(&restoreOnly, "only", "", "Comma-separated list of stages to run (exclusive with --skip)")
	restoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "Show what would be executed without doing it")
	rootCmd.AddCommand(restoreCmd)
}
