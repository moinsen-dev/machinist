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
	restoreList   bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore [manifest.toml]",
	Short: "Restore environment from manifest",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if restoreList {
			fmt.Fprintln(cmd.OutOrStdout(), "Available restore groups:")
			for _, g := range domain.RestoreGroups() {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", g.Name, g.Label)
			}
			return nil
		}

		manifestPath := "manifest.toml"
		if len(args) > 0 {
			manifestPath = args[0]
		}

		if restoreSkip != "" && restoreOnly != "" {
			return fmt.Errorf("--skip and --only are mutually exclusive; use one or the other")
		}

		snap, err := domain.ReadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("read manifest: %w", err)
		}

		// Build selected groups: all groups with data, filtered by --only/--skip
		allGroups := domain.RestoreGroups()
		var selected []domain.RestoreGroup

		if restoreOnly != "" {
			only := parseCSV(restoreOnly)
			validNames := make(map[string]bool)
			for _, n := range domain.GroupNames() {
				validNames[n] = true
			}
			var unknown []string
			for _, n := range only {
				if !validNames[n] {
					unknown = append(unknown, n)
				}
			}
			if len(unknown) > 0 {
				return fmt.Errorf("unknown group(s): %s (valid: %s)", strings.Join(unknown, ", "), strings.Join(domain.GroupNames(), ", "))
			}
			onlySet := make(map[string]bool, len(only))
			for _, name := range only {
				onlySet[name] = true
			}
			for _, g := range allGroups {
				if onlySet[g.Name] && g.HasData(snap) {
					selected = append(selected, g)
				}
			}
		} else if restoreSkip != "" {
			skip := parseCSV(restoreSkip)
			validNames := make(map[string]bool)
			for _, n := range domain.GroupNames() {
				validNames[n] = true
			}
			var unknown []string
			for _, n := range skip {
				if !validNames[n] {
					unknown = append(unknown, n)
				}
			}
			if len(unknown) > 0 {
				return fmt.Errorf("unknown group(s): %s (valid: %s)", strings.Join(unknown, ", "), strings.Join(domain.GroupNames(), ", "))
			}
			skipSet := make(map[string]bool, len(skip))
			for _, name := range skip {
				skipSet[name] = true
			}
			for _, g := range allGroups {
				if !skipSet[g.Name] && g.HasData(snap) {
					selected = append(selected, g)
				}
			}
		} else {
			for _, g := range allGroups {
				if g.HasData(snap) {
					selected = append(selected, g)
				}
			}
		}

		if restoreDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "Dry-run mode: restore plan\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Manifest: %s\n", manifestPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Host: %s (%s)\n", snap.Meta.SourceHostname, snap.Meta.SourceArch)
			fmt.Fprintf(cmd.OutOrStdout(), "Groups to execute: %d\n", len(selected))
			for i, g := range selected {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s (%d stages)\n", i+1, g.Name, g.StageCount(snap))
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nNo changes were made (dry-run).")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Restoring %d groups from %s\n", len(selected), manifestPath)
		for _, g := range selected {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", g.Name)
		}

		if !restoreYes {
			fmt.Fprintln(cmd.OutOrStdout(), "\nUse --yes to confirm execution.")
			return nil
		}

		// Execute: for each selected group, find script in bundle dir or generate on-the-fly
		bundleDir := filepath.Dir(manifestPath)

		// Try to use pre-existing group scripts from the bundle
		// Fall back to generating them on-the-fly
		scripts := make(map[string]string)
		for _, g := range selected {
			scriptPath := filepath.Join(bundleDir, g.ScriptName)
			if _, statErr := os.Stat(scriptPath); os.IsNotExist(statErr) {
				// Need to generate scripts on-the-fly
				if len(scripts) == 0 {
					scripts, err = bundler.GenerateRestoreScripts(snap)
					if err != nil {
						return fmt.Errorf("generate restore scripts: %w", err)
					}
				}
			}
		}

		for _, g := range selected {
			scriptPath := filepath.Join(bundleDir, g.ScriptName)

			// If the script doesn't exist on disk, write generated version to temp
			if _, statErr := os.Stat(scriptPath); os.IsNotExist(statErr) {
				content, ok := scripts[g.ScriptName]
				if !ok {
					fmt.Fprintf(cmd.ErrOrStderr(), "Warning: no script for group %s, skipping\n", g.Name)
					continue
				}
				tmpPath := filepath.Join(os.TempDir(), "machinist-"+g.ScriptName)
				if writeErr := os.WriteFile(tmpPath, []byte(content), 0755); writeErr != nil {
					return fmt.Errorf("write temp script %s: %w", g.ScriptName, writeErr)
				}
				defer os.Remove(tmpPath)
				scriptPath = tmpPath
			}

			fmt.Fprintf(cmd.OutOrStdout(), "==> Running %s ...\n", g.Name)
			execCmd := exec.CommandContext(cmd.Context(), "bash", scriptPath)
			execCmd.Dir = bundleDir
			execCmd.Stdout = cmd.OutOrStdout()
			execCmd.Stderr = cmd.ErrOrStderr()
			if runErr := execCmd.Run(); runErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: group %s failed: %v\n", g.Name, runErr)
			}
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

func init() {
	restoreCmd.Flags().StringVar(&restoreSkip, "skip", "", "Comma-separated list of groups to skip")
	restoreCmd.Flags().StringVar(&restoreOnly, "only", "", "Comma-separated list of groups to run (exclusive with --skip)")
	restoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "Show what would be executed without doing it")
	restoreCmd.Flags().BoolVarP(&restoreYes, "yes", "y", false, "Skip confirmation prompt")
	restoreCmd.Flags().BoolVar(&restoreList, "list", false, "List available restore groups")
	rootCmd.AddCommand(restoreCmd)
}
