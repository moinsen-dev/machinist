package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// AzureScanner scans Azure CLI configuration.
type AzureScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewAzureScanner creates a new AzureScanner for the given homeDir.
func NewAzureScanner(homeDir string, cmd util.CommandRunner) *AzureScanner {
	return &AzureScanner{homeDir: homeDir, cmd: cmd}
}

func (s *AzureScanner) Name() string        { return "azure" }
func (s *AzureScanner) Description() string  { return "Scans Azure CLI configuration" }
func (s *AzureScanner) Category() string     { return "cloud" }

// Scan checks for an Azure CLI installation and its configuration directory.
func (s *AzureScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	if !s.cmd.IsInstalled(ctx, "az") {
		return result, nil
	}

	section := &domain.AzureSection{}

	configDir := filepath.Join(s.homeDir, ".azure")
	if util.DirExists(configDir) {
		section.ConfigDir = ".azure"
	}

	if section.ConfigDir != "" {
		result.Azure = section
	}
	return result, nil
}
