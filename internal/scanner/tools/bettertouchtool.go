package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// BetterTouchToolScanner scans BetterTouchTool configuration.
type BetterTouchToolScanner struct {
	homeDir string
}

// NewBetterTouchToolScanner creates a new BetterTouchToolScanner for the given homeDir.
func NewBetterTouchToolScanner(homeDir string) *BetterTouchToolScanner {
	return &BetterTouchToolScanner{homeDir: homeDir}
}

func (s *BetterTouchToolScanner) Name() string        { return "bettertouchtool" }
func (s *BetterTouchToolScanner) Description() string  { return "Scans BetterTouchTool configuration" }
func (s *BetterTouchToolScanner) Category() string     { return "tools" }

// Scan checks for the BetterTouchTool application support directory.
func (s *BetterTouchToolScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	configDir := filepath.Join(s.homeDir, "Library", "Application Support", "BetterTouchTool")
	if !util.DirExists(configDir) {
		return result, nil
	}

	result.BetterTouchTool = &domain.BetterTouchToolSection{
		ConfigFile: filepath.Join("Library", "Application Support", "BetterTouchTool"),
	}
	return result, nil
}
