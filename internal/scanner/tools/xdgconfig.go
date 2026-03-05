package tools

import (
	"context"
	"os"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// knownXDGTools is the set of tools to auto-detect under ~/.config/.
var knownXDGTools = map[string]bool{
	"bat":       true,
	"lazygit":   true,
	"htop":      true,
	"btop":      true,
	"aerospace": true,
	"gh":        true,
	"starship":  true,
	"fish":      true,
	"kitty":     true,
	"micro":     true,
}

// XDGConfigScanner scans the ~/.config/ directory for known tool configurations.
type XDGConfigScanner struct {
	homeDir string
}

// NewXDGConfigScanner creates a new XDGConfigScanner with the given home directory.
func NewXDGConfigScanner(homeDir string) *XDGConfigScanner {
	return &XDGConfigScanner{homeDir: homeDir}
}

func (s *XDGConfigScanner) Name() string        { return "xdg-config" }
func (s *XDGConfigScanner) Description() string  { return "Scans XDG config directory for known tool configurations" }
func (s *XDGConfigScanner) Category() string     { return "tools" }

// Scan walks ~/.config/ one level deep looking for known tool directories.
func (s *XDGConfigScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	configDir := filepath.Join(s.homeDir, ".config")
	if !util.DirExists(configDir) {
		return result, nil
	}

	entries, err := os.ReadDir(configDir)
	if err != nil {
		return result, nil
	}

	section := &domain.XDGConfigSection{
		ConfigDir: configDir,
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if knownXDGTools[name] {
			section.AutoDetected = append(section.AutoDetected, name)
		}
	}

	result.XDGConfig = section
	return result, nil
}
