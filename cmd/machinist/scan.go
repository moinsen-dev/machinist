package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	gitscanner "github.com/moinsen-dev/machinist/internal/scanner/git"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/spf13/cobra"
)

var scanSearchPaths string

var scanCmd = &cobra.Command{
	Use:   "scan [scanner-name]",
	Short: "Run a single scanner",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg := newRegistry()
		ctx := context.Background()

		var result *scanner.ScanResult
		var err error

		if scanSearchPaths != "" && args[0] == "git-repos" {
			paths := strings.Split(scanSearchPaths, ",")
			for i := range paths {
				paths[i] = strings.TrimSpace(paths[i])
			}
			cmdRunner := &util.RealCommandRunner{}
			customReg := scanner.NewRegistry()
			customReg.Register(gitscanner.NewGitReposScanner(paths, cmdRunner))
			result, err = customReg.ScanOne(ctx, args[0])
		} else {
			result, err = reg.ScanOne(ctx, args[0])
		}

		if err != nil {
			return err
		}

		snap := domain.NewSnapshot("", "", "", Version)
		scanner.ApplyResult(snap, result)

		data, err := domain.MarshalManifest(snap)
		if err != nil {
			return fmt.Errorf("marshal: %w", err)
		}

		cmd.Print(string(data))
		return nil
	},
}

func init() {
	scanCmd.Flags().StringVar(&scanSearchPaths, "search-paths", "", "Comma-separated search paths for git-repos scanner")
	rootCmd.AddCommand(scanCmd)
}
