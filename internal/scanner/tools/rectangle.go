package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// RectangleScanner scans Rectangle window manager configuration.
type RectangleScanner struct {
	homeDir string
}

// NewRectangleScanner creates a new RectangleScanner for the given homeDir.
func NewRectangleScanner(homeDir string) *RectangleScanner {
	return &RectangleScanner{homeDir: homeDir}
}

func (s *RectangleScanner) Name() string        { return "rectangle" }
func (s *RectangleScanner) Description() string  { return "Scans Rectangle window manager configuration" }
func (s *RectangleScanner) Category() string     { return "tools" }

// Scan checks for the Rectangle preferences plist.
func (s *RectangleScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	configFile := filepath.Join(s.homeDir, "Library", "Preferences", "com.knewton.Rectangle.plist")
	if !util.FileExists(configFile) {
		return result, nil
	}

	result.Rectangle = &domain.RectangleSection{
		ConfigFile: configFile,
	}
	return result, nil
}
