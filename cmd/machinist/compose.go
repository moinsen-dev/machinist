package main

import (
	"fmt"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/profiles"
	"github.com/spf13/cobra"
)

var (
	composeFrom   string
	composeOutput string
	composeAdd    string
)

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Compose a manifest from a profile",
	Long:  "Build a setup manifest starting from a built-in profile, optionally adding extra packages.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if composeFrom == "" {
			available, _ := profiles.List()
			fmt.Fprintln(cmd.ErrOrStderr(), "Error: --from is required")
			fmt.Fprintln(cmd.ErrOrStderr(), "Available profiles:", strings.Join(available, ", "))
			return fmt.Errorf("--from flag is required")
		}

		// Strip profile:// prefix if present
		name := strings.TrimPrefix(composeFrom, "profile://")

		snap, err := profiles.Get(name)
		if err != nil {
			return fmt.Errorf("load profile %q: %w", name, err)
		}

		// Add extra packages
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
	composeCmd.Flags().StringVar(&composeFrom, "from", "", "Profile name (e.g. minimal, fullstack-js) or profile://name")
	composeCmd.Flags().StringVarP(&composeOutput, "output", "o", "composed-manifest.toml", "Output file path")
	composeCmd.Flags().StringVar(&composeAdd, "add", "", "Comma-separated packages to add (e.g. docker,postgres)")
	rootCmd.AddCommand(composeCmd)
}
