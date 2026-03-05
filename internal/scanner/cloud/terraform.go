package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// TerraformScanner scans Terraform CLI configuration.
type TerraformScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewTerraformScanner creates a new TerraformScanner with the given home directory and CommandRunner.
func NewTerraformScanner(homeDir string, cmd util.CommandRunner) *TerraformScanner {
	return &TerraformScanner{homeDir: homeDir, cmd: cmd}
}

func (s *TerraformScanner) Name() string        { return "terraform" }
func (s *TerraformScanner) Description() string { return "Scans Terraform CLI configuration" }
func (s *TerraformScanner) Category() string    { return "cloud" }

// Scan checks for Terraform installation and configuration files.
func (s *TerraformScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	if !s.cmd.IsInstalled(ctx, "terraform") {
		return result, nil
	}

	section := &domain.TerraformSection{}

	// Check for ~/.terraformrc
	rcPath := filepath.Join(s.homeDir, ".terraformrc")
	if util.FileExists(rcPath) {
		section.ConfigFile = ".terraformrc"
	}

	// Also check for ~/.terraform.d/ directory
	tfDir := filepath.Join(s.homeDir, ".terraform.d")
	if util.DirExists(tfDir) && section.ConfigFile == "" {
		section.ConfigFile = ".terraform.d/"
	}

	result.Terraform = section
	return result, nil
}
