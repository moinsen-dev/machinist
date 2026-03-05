package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// AWSScanner scans AWS CLI configuration and profiles.
type AWSScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewAWSScanner creates a new AWSScanner with the given home directory and CommandRunner.
func NewAWSScanner(homeDir string, cmd util.CommandRunner) *AWSScanner {
	return &AWSScanner{homeDir: homeDir, cmd: cmd}
}

func (s *AWSScanner) Name() string        { return "aws" }
func (s *AWSScanner) Description() string { return "Scans AWS CLI configuration and profiles" }
func (s *AWSScanner) Category() string    { return "cloud" }

// Scan checks for AWS CLI installation, config file, and configured profiles.
func (s *AWSScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	if !s.cmd.IsInstalled(ctx, "aws") {
		return result, nil
	}

	section := &domain.AWSSection{}

	// Check for ~/.aws/config
	configPath := filepath.Join(s.homeDir, ".aws", "config")
	if util.FileExists(configPath) {
		section.ConfigFile = ".aws/config"
	}

	// List profiles
	profiles, err := s.cmd.RunLines(ctx, "aws", "configure", "list-profiles")
	if err == nil {
		section.Profiles = profiles
	}

	result.AWS = section
	return result, nil
}
