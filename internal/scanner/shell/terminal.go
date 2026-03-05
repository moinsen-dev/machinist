package shell

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// TerminalScanner scans for terminal emulator configuration files.
type TerminalScanner struct {
	homeDir string
}

// NewTerminalScanner creates a new TerminalScanner that scans the given homeDir.
func NewTerminalScanner(homeDir string) *TerminalScanner {
	return &TerminalScanner{homeDir: homeDir}
}

func (t *TerminalScanner) Name() string        { return "terminal" }
func (t *TerminalScanner) Description() string  { return "Scans terminal emulator configuration" }
func (t *TerminalScanner) Category() string     { return "shell" }

// terminalCandidate describes a terminal emulator and how to detect it.
type terminalCandidate struct {
	app        string
	configPath string // relative to homeDir; empty means use the path directly
	isDir      bool   // detect via DirExists instead of FileExists
}

// Scan checks for well-known terminal emulator config files/dirs.
// The first match wins. Returns a nil Terminal section if none found.
func (t *TerminalScanner) Scan(_ context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: t.Name(),
	}

	candidates := []terminalCandidate{
		{
			app:        "iTerm2",
			configPath: "Library/Preferences/com.googlecode.iterm2.plist",
			isDir:      false,
		},
		{
			app:        "Warp",
			configPath: ".warp",
			isDir:      true,
		},
		{
			app:        "Alacritty",
			configPath: ".config/alacritty/alacritty.toml",
			isDir:      false,
		},
		{
			app:        "Alacritty",
			configPath: ".config/alacritty/alacritty.yml",
			isDir:      false,
		},
		{
			app:        "WezTerm",
			configPath: ".config/wezterm/wezterm.lua",
			isDir:      false,
		},
	}

	for _, c := range candidates {
		absPath := filepath.Join(t.homeDir, c.configPath)
		found := false
		if c.isDir {
			found = util.DirExists(absPath)
		} else {
			found = util.FileExists(absPath)
		}
		if !found {
			continue
		}

		section := &domain.TerminalSection{
			App: c.app,
		}

		// Record the config file/dir. Dirs are captured without a hash.
		if !c.isDir {
			hash, err := util.ContentHash(absPath)
			if err == nil {
				section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
					Source:      c.configPath,
					BundlePath:  filepath.Join("configs", "terminal", filepath.Base(c.configPath)),
					ContentHash: hash,
				})
			} else {
				section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
					Source:     c.configPath,
					BundlePath: filepath.Join("configs", "terminal", filepath.Base(c.configPath)),
				})
			}
		} else {
			section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
				Source:     c.configPath,
				BundlePath: filepath.Join("configs", "terminal", filepath.Base(c.configPath)),
			})
		}

		result.Terminal = section
		return result, nil
	}

	// No terminal emulator detected.
	return result, nil
}
