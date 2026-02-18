package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/bundler"
	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/spf13/cobra"
)

var (
	restoreSkip   string
	restoreOnly   string
	restoreDryRun bool
	restoreYes    bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore <manifest.toml>",
	Short: "Restore environment from manifest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manifestPath := args[0]

		if restoreSkip != "" && restoreOnly != "" {
			return fmt.Errorf("--skip and --only are mutually exclusive; use one or the other")
		}

		snap, err := domain.ReadManifest(manifestPath)
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
			fmt.Fprintf(cmd.OutOrStdout(), "Manifest: %s\n", manifestPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Host: %s (%s)\n", snap.Meta.SourceHostname, snap.Meta.SourceArch)
			fmt.Fprintf(cmd.OutOrStdout(), "Stages to execute: %d\n", len(sections))
			for i, s := range sections {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s\n", i+1, s)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nNo changes were made (dry-run).")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Restoring %d stages from %s\n", len(sections), manifestPath)
		for _, s := range sections {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", s)
		}

		if !restoreYes {
			fmt.Fprintln(cmd.OutOrStdout(), "\nUse --yes to confirm execution.")
			return nil
		}

		// Find or generate the install script
		scriptPath := filepath.Join(filepath.Dir(manifestPath), "install.command")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			script, genErr := bundler.GenerateRestoreScript(snap)
			if genErr != nil {
				return fmt.Errorf("generate restore script: %w", genErr)
			}
			scriptPath = filepath.Join(os.TempDir(), "machinist-restore.sh")
			if writeErr := os.WriteFile(scriptPath, []byte(script), 0755); writeErr != nil {
				return fmt.Errorf("write temp script: %w", writeErr)
			}
			defer os.Remove(scriptPath)
		}

		execCmd := exec.CommandContext(cmd.Context(), "bash", scriptPath)
		execCmd.Dir = filepath.Dir(manifestPath)
		execCmd.Stdout = cmd.OutOrStdout()
		execCmd.Stderr = cmd.ErrOrStderr()
		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("restore script failed: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "\nRestore complete.")
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
	restoreCmd.Flags().StringVar(&restoreSkip, "skip", "", "Comma-separated list of stages to skip")
	restoreCmd.Flags().StringVar(&restoreOnly, "only", "", "Comma-separated list of stages to run (exclusive with --skip)")
	restoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "Show what would be executed without doing it")
	restoreCmd.Flags().BoolVarP(&restoreYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(restoreCmd)
}
