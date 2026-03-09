package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// RaycastScanner scans Raycast configuration.
type RaycastScanner struct {
	homeDir string
}

// NewRaycastScanner creates a new RaycastScanner for the given homeDir.
func NewRaycastScanner(homeDir string) *RaycastScanner {
	return &RaycastScanner{homeDir: homeDir}
}

func (s *RaycastScanner) Name() string        { return "raycast" }
func (s *RaycastScanner) Description() string  { return "Scans Raycast launcher configuration" }
func (s *RaycastScanner) Category() string     { return "tools" }

// Scan checks for a Raycast application support directory.
func (s *RaycastScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	exportDir := filepath.Join(s.homeDir, "Library", "Application Support", "com.raycast.macos")
	if !util.DirExists(exportDir) {
		return result, nil
	}

	result.Raycast = &domain.RaycastSection{
		ExportFile: filepath.Join("Library", "Application Support", "com.raycast.macos"),
	}
	return result, nil
}
