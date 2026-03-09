package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// FlyioScanner scans Fly.io CLI configuration.
type FlyioScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewFlyioScanner creates a new FlyioScanner for the given homeDir.
func NewFlyioScanner(homeDir string, cmd util.CommandRunner) *FlyioScanner {
	return &FlyioScanner{homeDir: homeDir, cmd: cmd}
}

func (s *FlyioScanner) Name() string        { return "flyio" }
func (s *FlyioScanner) Description() string  { return "Scans Fly.io CLI configuration" }
func (s *FlyioScanner) Category() string     { return "cloud" }

// Scan checks for a Fly.io CLI installation and its configuration file.
func (s *FlyioScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	if !s.cmd.IsInstalled(ctx, "fly") {
		return result, nil
	}

	section := &domain.FlyioSection{}

	configFile := filepath.Join(s.homeDir, ".fly", "config.yml")
	if util.FileExists(configFile) {
		section.ConfigFile = filepath.Join(".fly", "config.yml")
	}

	if section.ConfigFile != "" {
		result.Flyio = section
	}
	return result, nil
}
