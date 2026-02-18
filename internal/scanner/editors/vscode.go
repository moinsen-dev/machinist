package editors

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// ---------------------------------------------------------------------------
// VSCodeScanner
// ---------------------------------------------------------------------------

// VSCodeScanner scans VS Code extensions, configuration files, and snippets.
type VSCodeScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewVSCodeScanner creates a new VSCodeScanner that scans the given homeDir.
func NewVSCodeScanner(homeDir string, cmd util.CommandRunner) *VSCodeScanner {
	return &VSCodeScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (s *VSCodeScanner) Name() string        { return "vscode" }
func (s *VSCodeScanner) Description() string { return "Scans VS Code extensions, settings, and snippets" }
func (s *VSCodeScanner) Category() string    { return "editors" }

// Scan inspects VS Code extensions, config files, and snippets directory.
func (s *VSCodeScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	section := &domain.VSCodeSection{}
	found := false

	// 1. Try to list extensions via the code CLI.
	extensions, err := s.cmd.RunLines(ctx, "code", "--list-extensions")
	if err == nil {
		found = true
		section.Extensions = extensions
	}

	// 2. Check config directory.
	configDir := filepath.Join(s.homeDir, "Library", "Application Support", "Code", "User")

	// 3. Check for settings.json and keybindings.json.
	for _, name := range []string{"settings.json", "keybindings.json"} {
		absPath := filepath.Join(configDir, name)
		if !util.FileExists(absPath) {
			continue
		}
		found = true
		hash, hashErr := util.ContentHash(absPath)
		if hashErr != nil {
			continue
		}
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:      "Library/Application Support/Code/User/" + name,
			BundlePath:  filepath.Join("configs", "vscode", name),
			ContentHash: hash,
		})
	}

	// 4. Check for snippets directory.
	snippetsPath := filepath.Join(configDir, "snippets")
	if util.DirExists(snippetsPath) {
		found = true
		section.SnippetsDir = "Library/Application Support/Code/User/snippets"
	}

	if !found {
		return &scanner.ScanResult{ScannerName: s.Name()}, nil
	}

	return &scanner.ScanResult{
		ScannerName: s.Name(),
		VSCode:      section,
	}, nil
}

// ---------------------------------------------------------------------------
// CursorScanner
// ---------------------------------------------------------------------------

// CursorScanner scans Cursor editor extensions and configuration files.
type CursorScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewCursorScanner creates a new CursorScanner that scans the given homeDir.
func NewCursorScanner(homeDir string, cmd util.CommandRunner) *CursorScanner {
	return &CursorScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (s *CursorScanner) Name() string        { return "cursor" }
func (s *CursorScanner) Description() string { return "Scans Cursor editor extensions and settings" }
func (s *CursorScanner) Category() string    { return "editors" }

// Scan inspects Cursor extensions and config files.
func (s *CursorScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	section := &domain.CursorSection{}
	found := false

	// 1. Try to list extensions via the cursor CLI.
	extensions, err := s.cmd.RunLines(ctx, "cursor", "--list-extensions")
	if err == nil {
		found = true
		section.Extensions = extensions
	}

	// 2. Check config directory for settings.json.
	configDir := filepath.Join(s.homeDir, "Library", "Application Support", "Cursor", "User")
	settingsPath := filepath.Join(configDir, "settings.json")
	if util.FileExists(settingsPath) {
		found = true
		hash, hashErr := util.ContentHash(settingsPath)
		if hashErr == nil {
			section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
				Source:      "Library/Application Support/Cursor/User/settings.json",
				BundlePath:  "configs/cursor/settings.json",
				ContentHash: hash,
			})
		}
	}

	if !found {
		return &scanner.ScanResult{ScannerName: s.Name()}, nil
	}

	return &scanner.ScanResult{
		ScannerName: s.Name(),
		Cursor:      section,
	}, nil
}
