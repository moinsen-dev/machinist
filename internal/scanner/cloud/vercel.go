package cloud

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// VercelScanner scans Vercel CLI configuration.
type VercelScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewVercelScanner creates a new VercelScanner for the given homeDir.
func NewVercelScanner(homeDir string, cmd util.CommandRunner) *VercelScanner {
	return &VercelScanner{homeDir: homeDir, cmd: cmd}
}

func (s *VercelScanner) Name() string        { return "vercel" }
func (s *VercelScanner) Description() string  { return "Scans Vercel CLI configuration" }
func (s *VercelScanner) Category() string     { return "cloud" }

// Scan checks for a Vercel CLI installation and its configuration directory.
func (s *VercelScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	if !s.cmd.IsInstalled(ctx, "vercel") {
		return result, nil
	}

	configDir := filepath.Join(s.homeDir, ".vercel")
	if util.DirExists(configDir) {
		result.Vercel = &domain.VercelSection{
			ConfigDir: ".vercel",
		}
	}

	return result, nil
}
