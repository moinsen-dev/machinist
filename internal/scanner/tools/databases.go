package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// DatabasesScanner scans database client configuration files.
type DatabasesScanner struct {
	homeDir string
}

// NewDatabasesScanner creates a new DatabasesScanner.
func NewDatabasesScanner(homeDir string) *DatabasesScanner {
	return &DatabasesScanner{homeDir: homeDir}
}

func (s *DatabasesScanner) Name() string        { return "databases" }
func (s *DatabasesScanner) Description() string  { return "Scans database client configuration files" }
func (s *DatabasesScanner) Category() string     { return "tools" }

// Scan checks for database configuration files and directories.
func (s *DatabasesScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	type dbEntry struct {
		path       string
		isDir      bool
		source     string
		bundlePath string
		sensitive  bool
	}

	entries := []dbEntry{
		{
			path:       filepath.Join(s.homeDir, ".pgpass"),
			isDir:      false,
			source:     ".pgpass",
			bundlePath: "configs/.pgpass",
			sensitive:  true,
		},
		{
			path:       filepath.Join(s.homeDir, ".my.cnf"),
			isDir:      false,
			source:     ".my.cnf",
			bundlePath: "configs/.my.cnf",
			sensitive:  true,
		},
		{
			path:       filepath.Join(s.homeDir, "Library", "Application Support", "com.tinyapp.TablePlus"),
			isDir:      true,
			source:     "TablePlus",
			bundlePath: "configs/TablePlus",
			sensitive:  false,
		},
		{
			path:       filepath.Join(s.homeDir, ".dbeaver4"),
			isDir:      true,
			source:     "DBeaver",
			bundlePath: "configs/.dbeaver4",
			sensitive:  false,
		},
	}

	var configFiles []domain.ConfigFile
	for _, e := range entries {
		exists := false
		if e.isDir {
			exists = util.DirExists(e.path)
		} else {
			exists = util.FileExists(e.path)
		}
		if exists {
			configFiles = append(configFiles, domain.ConfigFile{
				Source:     e.source,
				BundlePath: e.bundlePath,
				Sensitive:  e.sensitive,
			})
		}
	}

	if len(configFiles) > 0 {
		result.Databases = &domain.DatabasesSection{
			ConfigFiles: configFiles,
		}
	}

	return result, nil
}
