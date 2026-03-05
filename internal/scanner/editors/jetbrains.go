package editors

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
)

// jetBrainsIDENames maps known JetBrains settings directory prefixes to their display names.
var jetBrainsIDENames = map[string]string{
	"IntelliJIdea":  "IntelliJ IDEA",
	"PyCharm":       "PyCharm Professional",
	"WebStorm":      "WebStorm",
	"GoLand":        "GoLand",
	"AndroidStudio": "Android Studio",
	"PhpStorm":      "PhpStorm",
	"CLion":         "CLion",
	"Rider":         "Rider",
	"DataGrip":      "DataGrip",
	"RubyMine":      "RubyMine",
}

// JetBrainsScanner scans JetBrains IDE installations by inspecting the
// Application Support directory for known IDE settings directories.
type JetBrainsScanner struct {
	homeDir string
}

// NewJetBrainsScanner creates a new JetBrainsScanner for the given homeDir.
func NewJetBrainsScanner(homeDir string) *JetBrainsScanner {
	return &JetBrainsScanner{homeDir: homeDir}
}

func (s *JetBrainsScanner) Name() string        { return "jetbrains" }
func (s *JetBrainsScanner) Description() string { return "Scans JetBrains IDE installations" }
func (s *JetBrainsScanner) Category() string    { return "editors" }

// Scan walks ~/Library/Application Support/JetBrains/ and detects installed IDEs
// by matching subdirectory names against known IDE prefixes.
func (s *JetBrainsScanner) Scan(_ context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	jetbrainsDir := filepath.Join(s.homeDir, "Library", "Application Support", "JetBrains")
	entries, err := os.ReadDir(jetbrainsDir)
	if err != nil {
		// Directory doesn't exist or isn't readable — treat as not installed.
		return result, nil
	}

	var ides []domain.JetBrainsIDE
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		displayName := matchJetBrainsPrefix(name)
		if displayName == "" {
			continue
		}
		ides = append(ides, domain.JetBrainsIDE{
			Name:           displayName,
			SettingsExport: filepath.Join(jetbrainsDir, name),
		})
	}

	if len(ides) == 0 {
		return result, nil
	}

	result.JetBrains = &domain.JetBrainsSection{IDEs: ides}
	return result, nil
}

// matchJetBrainsPrefix returns the display name for a JetBrains settings directory
// (e.g. "IntelliJIdea2023.3" → "IntelliJ IDEA"), or "" if not recognised.
func matchJetBrainsPrefix(dirName string) string {
	for prefix, displayName := range jetBrainsIDENames {
		if strings.HasPrefix(dirName, prefix) {
			return displayName
		}
	}
	return ""
}
