package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/moinsen-dev/machinist/internal/bundler"
	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/tui"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/spf13/cobra"
)

var (
	dmgOutput      string
	dmgPassword    string
	dmgInteractive bool
)

var dmgCmd = &cobra.Command{
	Use:   "dmg [manifest.toml]",
	Short: "Create a DMG restore bundle from a manifest or by scanning the environment",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		var snap *domain.Snapshot
		var errs []error

		if len(args) == 1 {
			// Load from existing manifest file
			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read manifest: %w", err)
			}
			snap, err = domain.UnmarshalManifest(data)
			if err != nil {
				return fmt.Errorf("parse manifest: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Loaded manifest from %s (%d stages)\n", args[0], snap.StageCount())
		} else if dmgInteractive {
			reg := newRegistry()
			var err error
			snap, errs, err = runInteractiveScan(cmd, reg, ctx)
			if err != nil {
				return err
			}
			if snap == nil {
				return nil // cancelled
			}
		} else {
			reg := newRegistry()
			fmt.Fprintln(cmd.OutOrStdout(), "Scanning environment...")
			progress := newProgressWriter(cmd.OutOrStdout())
			snap, errs = reg.ScanAllWithProgress(ctx, progress)
		}

		for _, e := range errs {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", e)
		}

		// Prompt for age encryption passphrase if snapshot has encrypted sections
		var passphrase string
		if hasEncryptedSections(snap) {
			fmt.Fprintf(cmd.OutOrStdout(), "\nSensitive files detected (SSH keys, GPG, .env). These will be encrypted in the bundle.")
			fmt.Fprintf(cmd.OutOrStdout(), "\nEnter encryption passphrase (empty to skip encryption): ")
			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			passphrase = strings.TrimSpace(line)
			if passphrase == "" {
				fmt.Fprintf(cmd.OutOrStdout(), "\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "WARNING: Skipping encryption — the following will NOT be included in the DMG:\n")
			if snap.SSH != nil && len(snap.SSH.Keys) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "  - %d SSH key(s): %s\n", len(snap.SSH.Keys), strings.Join(snap.SSH.Keys, ", "))
			}
			if snap.GPG != nil && len(snap.GPG.ConfigFiles) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "  - %d GPG config file(s)\n", len(snap.GPG.ConfigFiles))
			}
			if snap.EnvFiles != nil && len(snap.EnvFiles.Files) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "  - %d .env file(s)\n", len(snap.EnvFiles.Files))
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "Re-run with a passphrase to include these files.\n")
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nBuilding DMG bundle...")
		runner := &util.RealCommandRunner{}
		opts := bundler.BundleOptions{
			Password:   dmgPassword,
			Passphrase: passphrase,
			VolumeName: "Machinist Restore",
		}
		if err := bundler.Bundle(ctx, runner, snap, dmgOutput, opts); err != nil {
			return fmt.Errorf("create DMG bundle: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), " done\nDMG bundle written to %s (%.1fs)\n", dmgOutput, snap.Meta.ScanDurationSecs)
		return nil
	},
}

// runInteractiveScan presents a TUI for scanner selection and runs selected scanners.
func runInteractiveScan(cmd *cobra.Command, reg *scanner.Registry, ctx context.Context) (*domain.Snapshot, []error, error) {
	scanners := reg.List()

	items := make([]tui.ScannerItem, len(scanners))
	for i, s := range scanners {
		items[i] = tui.ScannerItem{
			Name:        s.Name(),
			Description: s.Description(),
			Category:    s.Category(),
		}
	}

	selectModel := tui.NewScannerSelectModel(items)
	p := tea.NewProgram(selectModel)
	finalModel, err := p.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("interactive selection: %w", err)
	}

	result := finalModel.(tui.ScannerSelectModel)
	if result.Quitted() {
		fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
		return nil, nil, nil
	}

	selected := result.Selected()
	if len(selected) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No scanners selected.")
		return nil, nil, nil
	}

	start := time.Now()
	hostname, _ := os.Hostname()
	snap := domain.NewSnapshot(hostname, runtime.GOOS, runtime.GOARCH, "")

	var errs []error
	for _, name := range selected {
		scanResult, scanErr := reg.ScanOne(ctx, name)
		if scanErr != nil {
			errs = append(errs, scanErr)
			continue
		}
		scanner.ApplyResult(snap, scanResult)
	}

	snap.Meta.ScanDurationSecs = time.Since(start).Seconds()
	return snap, errs, nil
}

// hasEncryptedSections returns true if the snapshot contains sections marked as encrypted
// (SSH, GPG, or EnvFiles with Encrypted: true), indicating sensitive files that should
// be age-encrypted in the bundle.
func hasEncryptedSections(snap *domain.Snapshot) bool {
	if snap.SSH != nil && snap.SSH.Encrypted {
		return true
	}
	if snap.GPG != nil && snap.GPG.Encrypted {
		return true
	}
	if snap.EnvFiles != nil && snap.EnvFiles.Encrypted {
		return true
	}
	return false
}

func init() {
	dmgCmd.Flags().StringVarP(&dmgOutput, "output", "o", "machinist.dmg", "Output DMG file path")
	dmgCmd.Flags().StringVar(&dmgPassword, "password", "", "Encrypt DMG with password")
	dmgCmd.Flags().BoolVarP(&dmgInteractive, "interactive", "i", false, "Interactively select scanners")
	rootCmd.AddCommand(dmgCmd)
}
