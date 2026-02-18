package system

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// fontExtensions lists the file extensions recognised as font files.
var fontExtensions = map[string]bool{
	".ttf":   true,
	".otf":   true,
	".ttc":   true,
	".woff":  true,
	".woff2": true,
}

// FontsScanner scans user-installed fonts and Homebrew font casks.
type FontsScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewFontsScanner creates a new FontsScanner with the given home directory and CommandRunner.
func NewFontsScanner(homeDir string, cmd util.CommandRunner) *FontsScanner {
	return &FontsScanner{homeDir: homeDir, cmd: cmd}
}

func (f *FontsScanner) Name() string        { return "fonts" }
func (f *FontsScanner) Description() string  { return "Scans user-installed fonts and Homebrew font casks" }
func (f *FontsScanner) Category() string     { return "system" }

// Scan walks ~/Library/Fonts for user-installed font files and queries Homebrew
// for font-* casks. It returns a ScanResult with the Fonts field populated.
func (f *FontsScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: f.Name(),
	}

	section := &domain.FontsSection{}

	// 1. Walk ~/Library/Fonts/ for user-installed font files.
	fontsDir := filepath.Join(f.homeDir, "Library", "Fonts")
	if entries, err := os.ReadDir(fontsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if !fontExtensions[ext] {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			section.CustomFonts = append(section.CustomFonts, domain.Font{
				Name:       name,
				BundlePath: filepath.Join(fontsDir, entry.Name()),
			})
		}
	}

	// 2. Check Homebrew casks for font-* entries.
	if f.cmd.IsInstalled(ctx, "brew") {
		casks, err := f.cmd.RunLines(ctx, "brew", "list", "--cask")
		if err == nil {
			for _, cask := range casks {
				if strings.HasPrefix(cask, "font-") {
					section.HomebrewFonts = append(section.HomebrewFonts, cask)
				}
			}
		}
	}

	result.Fonts = section
	return result, nil
}
