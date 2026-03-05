package main

import (
	"fmt"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/profiles"
	"github.com/spf13/cobra"
)

var (
	composeOutput   string
	composeAdd      string
	composeFromFile string
)

var composeCmd = &cobra.Command{
	Use:   "compose [profile]",
	Short: "Compose a manifest from a built-in profile or existing manifest file",
	Long:  "Build a setup manifest starting from a built-in profile or an existing manifest file, optionally adding extra packages.\nUse 'machinist list profiles' to see available profiles.",
	Args: func(cmd *cobra.Command, args []string) error {
		if composeFromFile != "" {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var snap *domain.Snapshot
		var err error

		if composeFromFile != "" {
			snap, err = domain.ReadManifest(composeFromFile)
			if err != nil {
				return fmt.Errorf("read manifest: %w", err)
			}
		} else {
			name := strings.TrimPrefix(args[0], "profile://")
			snap, err = profiles.Get(name)
			if err != nil {
				available, _ := profiles.List()
				return fmt.Errorf("unknown profile %q (available: %s)", name, strings.Join(available, ", "))
			}
		}

		if composeAdd != "" {
			extras := strings.Split(composeAdd, ",")
			if snap.Homebrew == nil {
				snap.Homebrew = &domain.HomebrewSection{}
			}
			for _, pkg := range extras {
				pkg = strings.TrimSpace(pkg)
				if pkg != "" {
					snap.Homebrew.Formulae = append(snap.Homebrew.Formulae, domain.Package{Name: pkg})
				}
			}
		}

		if err := domain.WriteManifest(snap, composeOutput); err != nil {
			return fmt.Errorf("write manifest: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Composed manifest written to %s\n", composeOutput)
		return nil
	},
}

func init() {
	composeCmd.Flags().StringVarP(&composeOutput, "output", "o", "composed-manifest.toml", "Output file path")
	composeCmd.Flags().StringVar(&composeAdd, "add", "", "Comma-separated packages to add (e.g. docker,postgres)")
	composeCmd.Flags().StringVar(&composeFromFile, "from-file", "", "Use an existing manifest file as base")
	rootCmd.AddCommand(composeCmd)
}
