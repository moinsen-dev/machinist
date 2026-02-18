package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
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
	Use:   "dmg",
	Short: "Scan environment and create a DMG restore bundle",
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := newRegistry()
		ctx := context.Background()

		var snap *domain.Snapshot
		var errs []error

		if dmgInteractive {
			var err error
			snap, errs, err = runInteractiveScan(cmd, reg, ctx)
			if err != nil {
				return err
			}
			if snap == nil {
				return nil // cancelled
			}
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "Scanning environment...")
			progress := newProgressWriter(cmd.OutOrStdout())
			snap, errs = reg.ScanAllWithProgress(ctx, progress)
		}

		for _, e := range errs {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", e)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "\nBuilding DMG bundle...")
		runner := &util.RealCommandRunner{}
		opts := bundler.BundleOptions{
			Password:   dmgPassword,
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
		applyResultToSnapshot(snap, scanResult)
	}

	snap.Meta.ScanDurationSecs = time.Since(start).Seconds()
	return snap, errs, nil
}

func init() {
	dmgCmd.Flags().StringVarP(&dmgOutput, "output", "o", "machinist.dmg", "Output DMG file path")
	dmgCmd.Flags().StringVar(&dmgPassword, "password", "", "Encrypt DMG with password")
	dmgCmd.Flags().BoolVarP(&dmgInteractive, "interactive", "i", false, "Interactively select scanners")
	rootCmd.AddCommand(dmgCmd)
}
