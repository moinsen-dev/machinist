package editors

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// NeovimScanner scans Neovim configuration and detects the plugin manager in use.
type NeovimScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewNeovimScanner creates a new NeovimScanner for the given homeDir.
func NewNeovimScanner(homeDir string, cmd util.CommandRunner) *NeovimScanner {
	return &NeovimScanner{homeDir: homeDir, cmd: cmd}
}

func (s *NeovimScanner) Name() string        { return "neovim" }
func (s *NeovimScanner) Description() string { return "Scans Neovim configuration" }
func (s *NeovimScanner) Category() string    { return "editors" }

// Scan checks for a Neovim installation and config directory, then identifies
// the plugin manager (lazy.nvim, packer, or vim-plug) if one is present.
func (s *NeovimScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	if !s.cmd.IsInstalled(ctx, "nvim") {
		return result, nil
	}

	section := &domain.NeovimSection{}

	// Config directory.
	configDir := filepath.Join(s.homeDir, ".config", "nvim")
	if util.DirExists(configDir) {
		section.ConfigDir = filepath.Join(".config", "nvim")
	}

	// Plugin manager detection — check well-known data directories.
	shareBase := filepath.Join(s.homeDir, ".local", "share", "nvim")
	switch {
	case util.DirExists(filepath.Join(shareBase, "lazy")):
		section.PluginManager = "lazy.nvim"
	case util.DirExists(filepath.Join(shareBase, "site", "pack", "packer")):
		section.PluginManager = "packer"
	case util.DirExists(filepath.Join(shareBase, "plugged")):
		section.PluginManager = "vim-plug"
	}

	result.Neovim = section
	return result, nil
}
