package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// AlfredScanner scans Alfred configuration.
type AlfredScanner struct {
	homeDir string
}

// NewAlfredScanner creates a new AlfredScanner for the given homeDir.
func NewAlfredScanner(homeDir string) *AlfredScanner {
	return &AlfredScanner{homeDir: homeDir}
}

func (s *AlfredScanner) Name() string        { return "alfred" }
func (s *AlfredScanner) Description() string  { return "Scans Alfred launcher configuration" }
func (s *AlfredScanner) Category() string     { return "tools" }

// Scan checks for Alfred configuration directory with a fallback to the preferences plist.
func (s *AlfredScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	// Primary location
	configDir := filepath.Join(s.homeDir, "Library", "Application Support", "Alfred")
	if util.DirExists(configDir) {
		result.Alfred = &domain.AlfredSection{
			ConfigDir: filepath.Join("Library", "Application Support", "Alfred"),
		}
		return result, nil
	}

	// Fallback: preferences plist
	plistPath := filepath.Join(s.homeDir, "Library", "Preferences", "com.runningwithcrayons.Alfred-Preferences-3.plist")
	if util.FileExists(plistPath) {
		result.Alfred = &domain.AlfredSection{
			ConfigDir: filepath.Join("Library", "Preferences", "com.runningwithcrayons.Alfred-Preferences-3.plist"),
		}
		return result, nil
	}

	return result, nil
}
