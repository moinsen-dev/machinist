package main

import (
	"os"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/scanner/packages"
	"github.com/moinsen-dev/machinist/internal/scanner/shell"
	"github.com/moinsen-dev/machinist/internal/util"
)

func newRegistry() *scanner.Registry {
	cmd := &util.RealCommandRunner{}
	homeDir, _ := os.UserHomeDir()

	reg := scanner.NewRegistry()
	reg.Register(packages.NewHomebrewScanner(cmd))
	reg.Register(shell.NewShellConfigScanner(homeDir, cmd))
	return reg
}

// applyResultToSnapshot maps a ScanResult onto a Snapshot (used by scan command).
func applyResultToSnapshot(snap *domain.Snapshot, result *scanner.ScanResult) {
	if result.Homebrew != nil {
		snap.Homebrew = result.Homebrew
	}
	if result.Shell != nil {
		snap.Shell = result.Shell
	}
}
