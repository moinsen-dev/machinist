package main

import (
	"os"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/scanner/editors"
	gitscanner "github.com/moinsen-dev/machinist/internal/scanner/git"
	"github.com/moinsen-dev/machinist/internal/scanner/packages"
	"github.com/moinsen-dev/machinist/internal/scanner/runtimes"
	"github.com/moinsen-dev/machinist/internal/scanner/shell"
	"github.com/moinsen-dev/machinist/internal/util"
)

func newRegistry() *scanner.Registry {
	cmd := &util.RealCommandRunner{}
	homeDir, _ := os.UserHomeDir()

	// Default git repo search paths
	searchPaths := []string{
		filepath.Join(homeDir, "Code"),
		filepath.Join(homeDir, "Projects"),
		filepath.Join(homeDir, "Developer"),
		filepath.Join(homeDir, "work"),
	}

	reg := scanner.NewRegistry()
	reg.Register(packages.NewHomebrewScanner(cmd))
	reg.Register(shell.NewShellConfigScanner(homeDir, cmd))
	reg.Register(gitscanner.NewGitReposScanner(searchPaths, cmd))
	reg.Register(runtimes.NewNodeScanner(homeDir, cmd))
	reg.Register(runtimes.NewPythonScanner(cmd))
	reg.Register(runtimes.NewRustScanner(cmd))
	reg.Register(editors.NewVSCodeScanner(homeDir, cmd))
	reg.Register(editors.NewCursorScanner(homeDir, cmd))
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
	if result.GitRepos != nil {
		snap.GitRepos = result.GitRepos
	}
	if result.Node != nil {
		snap.Node = result.Node
	}
	if result.Python != nil {
		snap.Python = result.Python
	}
	if result.Rust != nil {
		snap.Rust = result.Rust
	}
	if result.VSCode != nil {
		snap.VSCode = result.VSCode
	}
	if result.Cursor != nil {
		snap.Cursor = result.Cursor
	}
}
