package system

import (
	"context"
	"os"
	"sort"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
)

// systemDirs are macOS default directories that exist on every Mac and should be skipped.
var systemDirs = map[string]bool{
	"Library":      true,
	"Applications": true,
	"Public":       true,
	"Movies":       true,
	"Music":        true,
	"Pictures":     true,
	"Desktop":      true,
	"Documents":    true,
	"Downloads":    true,
}

// FoldersScanner scans the home directory for custom top-level folder structure.
type FoldersScanner struct {
	homeDir string
}

// NewFoldersScanner creates a new FoldersScanner that scans the given home directory.
func NewFoldersScanner(homeDir string) *FoldersScanner {
	return &FoldersScanner{homeDir: homeDir}
}

func (f *FoldersScanner) Name() string        { return "folders" }
func (f *FoldersScanner) Description() string  { return "Scans home directory folder structure" }
func (f *FoldersScanner) Category() string     { return "system" }

// Scan lists top-level directories in homeDir, skipping hidden dirs and known macOS system dirs.
func (f *FoldersScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: f.Name(),
	}

	entries, err := os.ReadDir(f.homeDir)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()

		// Skip hidden directories.
		if len(name) > 0 && name[0] == '.' {
			continue
		}

		// Skip known macOS system directories.
		if systemDirs[name] {
			continue
		}

		dirs = append(dirs, name)
	}

	sort.Strings(dirs)

	result.Folders = &domain.FoldersSection{
		Structure: dirs,
	}

	return result, nil
}
