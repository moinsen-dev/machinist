package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// OnePasswordScanner scans 1Password CLI configuration.
type OnePasswordScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewOnePasswordScanner creates a new OnePasswordScanner.
func NewOnePasswordScanner(homeDir string, cmd util.CommandRunner) *OnePasswordScanner {
	return &OnePasswordScanner{homeDir: homeDir, cmd: cmd}
}

func (s *OnePasswordScanner) Name() string        { return "1password" }
func (s *OnePasswordScanner) Description() string  { return "Scans 1Password CLI configuration" }
func (s *OnePasswordScanner) Category() string     { return "tools" }

// Scan checks for the 1Password CLI and its configuration directory.
func (s *OnePasswordScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	if !s.cmd.IsInstalled(ctx, "op") {
		return result, nil
	}

	configDir := filepath.Join(s.homeDir, ".config", "op")
	if !util.DirExists(configDir) {
		return result, nil
	}

	result.OnePassword = &domain.OnePasswordSection{
		ConfigDir: filepath.Join(".config", "op"),
	}
	return result, nil
}
