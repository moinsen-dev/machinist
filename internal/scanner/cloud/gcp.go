package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// GCPScanner scans Google Cloud CLI configuration.
type GCPScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewGCPScanner creates a new GCPScanner for the given homeDir.
func NewGCPScanner(homeDir string, cmd util.CommandRunner) *GCPScanner {
	return &GCPScanner{homeDir: homeDir, cmd: cmd}
}

func (s *GCPScanner) Name() string        { return "gcp" }
func (s *GCPScanner) Description() string  { return "Scans Google Cloud CLI configuration" }
func (s *GCPScanner) Category() string     { return "cloud" }

// Scan checks for a gcloud CLI installation and its configuration directory.
func (s *GCPScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	if !s.cmd.IsInstalled(ctx, "gcloud") {
		return result, nil
	}

	section := &domain.GCPSection{}

	configDir := filepath.Join(s.homeDir, ".config", "gcloud")
	if util.DirExists(configDir) {
		section.ConfigDir = filepath.Join(".config", "gcloud")
	}

	if section.ConfigDir != "" {
		result.GCP = section
	}
	return result, nil
}
