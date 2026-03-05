package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// RegistriesScanner scans package registry configuration files.
type RegistriesScanner struct {
	homeDir string
}

// NewRegistriesScanner creates a new RegistriesScanner.
func NewRegistriesScanner(homeDir string) *RegistriesScanner {
	return &RegistriesScanner{homeDir: homeDir}
}

func (s *RegistriesScanner) Name() string        { return "registries" }
func (s *RegistriesScanner) Description() string  { return "Scans package registry configuration files" }
func (s *RegistriesScanner) Category() string     { return "tools" }

// Scan checks for registry configuration files.
func (s *RegistriesScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	type regEntry struct {
		path       string
		source     string
		bundlePath string
		sensitive  bool
	}

	entries := []regEntry{
		{
			path:       filepath.Join(s.homeDir, ".npmrc"),
			source:     ".npmrc",
			bundlePath: "configs/.npmrc",
			sensitive:  true,
		},
		{
			path:       filepath.Join(s.homeDir, ".pip", "pip.conf"),
			source:     ".pip/pip.conf",
			bundlePath: "configs/.pip/pip.conf",
			sensitive:  false,
		},
		{
			path:       filepath.Join(s.homeDir, ".cargo", "config.toml"),
			source:     ".cargo/config.toml",
			bundlePath: "configs/.cargo/config.toml",
			sensitive:  false,
		},
		{
			path:       filepath.Join(s.homeDir, ".gemrc"),
			source:     ".gemrc",
			bundlePath: "configs/.gemrc",
			sensitive:  false,
		},
		{
			path:       filepath.Join(s.homeDir, ".cocoapods", "config.yaml"),
			source:     ".cocoapods/config.yaml",
			bundlePath: "configs/.cocoapods/config.yaml",
			sensitive:  false,
		},
	}

	var configFiles []domain.ConfigFile
	for _, e := range entries {
		if util.FileExists(e.path) {
			configFiles = append(configFiles, domain.ConfigFile{
				Source:     e.source,
				BundlePath: e.bundlePath,
				Sensitive:  e.sensitive,
			})
		}
	}

	if len(configFiles) > 0 {
		result.Registries = &domain.RegistriesSection{
			ConfigFiles: configFiles,
		}
	}

	return result, nil
}
