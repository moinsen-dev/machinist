package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// envFileNames is the set of environment file names to look for.
var envFileNames = map[string]bool{
	".env":             true,
	".env.local":       true,
	".env.production":  true,
}

// workspaceDirs is the list of common workspace directories to search.
var workspaceDirs = []string{
	"Code",
	"Projects",
	"Developer",
	"work",
}

// EnvFilesScanner scans workspace directories for .env files.
type EnvFilesScanner struct {
	homeDir string
}

// NewEnvFilesScanner creates a new EnvFilesScanner with the given home directory.
func NewEnvFilesScanner(homeDir string) *EnvFilesScanner {
	return &EnvFilesScanner{homeDir: homeDir}
}

func (s *EnvFilesScanner) Name() string        { return "env-files" }
func (s *EnvFilesScanner) Description() string  { return "Scans workspace directories for .env files" }
func (s *EnvFilesScanner) Category() string     { return "tools" }

// Scan walks workspace directories 2 levels deep looking for .env files.
func (s *EnvFilesScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	var files []domain.EnvFile

	for _, dir := range workspaceDirs {
		wsDir := filepath.Join(s.homeDir, dir)
		if !util.DirExists(wsDir) {
			continue
		}
		found := s.walkForEnvFiles(wsDir, 0, 2)
		files = append(files, found...)
	}

	if len(files) == 0 {
		return result, nil
	}

	result.EnvFiles = &domain.EnvFilesSection{
		Encrypted: true,
		Files:     files,
	}
	return result, nil
}

// walkForEnvFiles recursively walks directories up to maxDepth levels,
// skipping hidden directories, looking for env files.
func (s *EnvFilesScanner) walkForEnvFiles(dir string, currentDepth, maxDepth int) []domain.EnvFile {
	var files []domain.EnvFile

	if currentDepth > maxDepth {
		return files
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}

	for _, entry := range entries {
		name := entry.Name()

		if entry.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(name, ".") {
				continue
			}
			if currentDepth < maxDepth {
				subFiles := s.walkForEnvFiles(filepath.Join(dir, name), currentDepth+1, maxDepth)
				files = append(files, subFiles...)
			}
			continue
		}

		// Check if this is an env file we care about
		if envFileNames[name] {
			fullPath := filepath.Join(dir, name)
			files = append(files, domain.EnvFile{
				Source:     fullPath,
				BundlePath: filepath.Join("env_files", strings.TrimPrefix(fullPath, s.homeDir+"/")),
			})
		}
	}

	return files
}
