package tools

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// APIToolsScanner scans API development tools like Postman, Insomnia, and mkcert.
type APIToolsScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewAPIToolsScanner creates a new APIToolsScanner with the given home directory and command runner.
func NewAPIToolsScanner(homeDir string, cmd util.CommandRunner) *APIToolsScanner {
	return &APIToolsScanner{homeDir: homeDir, cmd: cmd}
}

func (s *APIToolsScanner) Name() string        { return "api-tools" }
func (s *APIToolsScanner) Description() string  { return "Scans API development tool configurations" }
func (s *APIToolsScanner) Category() string     { return "tools" }

// Scan checks for Postman and Insomnia config directories, and mkcert availability.
func (s *APIToolsScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	section := &domain.APIToolsSection{}
	found := false

	// Check Postman
	postmanDir := filepath.Join(s.homeDir, "Library", "Application Support", "Postman")
	if util.DirExists(postmanDir) {
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:     "Postman",
			BundlePath: "configs/Postman",
		})
		found = true
	}

	// Check Insomnia
	insomniaDir := filepath.Join(s.homeDir, "Library", "Application Support", "Insomnia")
	if util.DirExists(insomniaDir) {
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:     "Insomnia",
			BundlePath: "configs/Insomnia",
		})
		found = true
	}

	// Check mkcert
	if s.cmd.IsInstalled(ctx, "mkcert") {
		section.Mkcert = true
		found = true
	}

	if found {
		result.APITools = section
	}

	return result, nil
}
