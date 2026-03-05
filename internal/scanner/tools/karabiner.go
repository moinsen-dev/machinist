package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// KarabinerScanner scans Karabiner-Elements configuration.
type KarabinerScanner struct {
	homeDir string
}

// NewKarabinerScanner creates a new KarabinerScanner for the given homeDir.
func NewKarabinerScanner(homeDir string) *KarabinerScanner {
	return &KarabinerScanner{homeDir: homeDir}
}

func (s *KarabinerScanner) Name() string        { return "karabiner" }
func (s *KarabinerScanner) Description() string  { return "Scans Karabiner-Elements configuration" }
func (s *KarabinerScanner) Category() string     { return "tools" }

// Scan checks for a Karabiner-Elements configuration directory containing karabiner.json.
func (s *KarabinerScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	configDir := filepath.Join(s.homeDir, ".config", "karabiner")
	if !util.DirExists(configDir) {
		return result, nil
	}

	// Verify karabiner.json exists inside the config directory
	configFile := filepath.Join(configDir, "karabiner.json")
	if !util.FileExists(configFile) {
		return result, nil
	}

	result.Karabiner = &domain.KarabinerSection{
		ConfigDir: configDir,
	}
	return result, nil
}
