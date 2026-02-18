package shell

import (
	"context"
	"path/filepath"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// knownShellConfigs lists shell configuration files to look for relative to homeDir.
var knownShellConfigs = []string{
	".zshrc", ".zshenv", ".zprofile", ".zlogin", ".zlogout",
	".bashrc", ".bash_profile", ".bash_login", ".bash_logout", ".profile",
	".inputrc",
}

// ShellConfigScanner scans shell configuration files and detects frameworks/prompts.
type ShellConfigScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewShellConfigScanner creates a new ShellConfigScanner that scans the given homeDir.
func NewShellConfigScanner(homeDir string, cmd util.CommandRunner) *ShellConfigScanner {
	return &ShellConfigScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (s *ShellConfigScanner) Name() string        { return "shell" }
func (s *ShellConfigScanner) Description() string  { return "Scans shell configuration files and frameworks" }
func (s *ShellConfigScanner) Category() string     { return "shell" }

// Scan inspects the home directory for shell config files, frameworks, and prompt tools.
func (s *ShellConfigScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	section := &domain.ShellSection{}

	// 1. Collect known shell config files.
	for _, name := range knownShellConfigs {
		absPath := filepath.Join(s.homeDir, name)
		if !util.FileExists(absPath) {
			continue
		}
		hash, err := util.ContentHash(absPath)
		if err != nil {
			// Skip files we cannot hash.
			continue
		}
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:      name,
			BundlePath:  filepath.Join("configs", "shell", name),
			ContentHash: hash,
		})
	}

	// 2. Detect default shell via command runner.
	shell, err := s.cmd.Run(ctx, "sh", "-c", "echo $SHELL")
	if err == nil && shell != "" {
		section.DefaultShell = shell
	}

	// 3. Framework detection.
	if util.DirExists(filepath.Join(s.homeDir, ".oh-my-zsh")) {
		section.Framework = "oh-my-zsh"
	} else if util.DirExists(filepath.Join(s.homeDir, ".oh-my-bash")) {
		section.Framework = "oh-my-bash"
	} else if util.DirExists(filepath.Join(s.homeDir, ".zprezto")) {
		section.Framework = "prezto"
	}

	// 4. Prompt detection.
	starshipPath := filepath.Join(s.homeDir, ".config", "starship.toml")
	p10kPath := filepath.Join(s.homeDir, ".p10k.zsh")

	if util.FileExists(starshipPath) {
		section.Prompt = "starship"
		hash, err := util.ContentHash(starshipPath)
		if err == nil {
			section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
				Source:      ".config/starship.toml",
				BundlePath:  "configs/shell/starship.toml",
				ContentHash: hash,
			})
		}
	} else if util.FileExists(p10kPath) {
		section.Prompt = "p10k"
		hash, err := util.ContentHash(p10kPath)
		if err == nil {
			section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
				Source:      ".p10k.zsh",
				BundlePath:  "configs/shell/.p10k.zsh",
				ContentHash: hash,
			})
		}
	}

	return &scanner.ScanResult{
		ScannerName: s.Name(),
		Shell:       section,
	}, nil
}
