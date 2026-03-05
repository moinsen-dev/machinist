package git

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// GitHubCLIScanner scans GitHub CLI extensions and configuration.
type GitHubCLIScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewGitHubCLIScanner creates a new GitHubCLIScanner with the given homeDir and CommandRunner.
func NewGitHubCLIScanner(homeDir string, cmd util.CommandRunner) *GitHubCLIScanner {
	return &GitHubCLIScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (g *GitHubCLIScanner) Name() string        { return "github-cli" }
func (g *GitHubCLIScanner) Description() string { return "Scans GitHub CLI extensions" }
func (g *GitHubCLIScanner) Category() string    { return "git" }

// Scan checks for gh CLI installation, reads extensions, and returns a ScanResult
// with the GitHubCLI field populated.
func (g *GitHubCLIScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: g.Name(),
	}

	if !g.cmd.IsInstalled(ctx, "gh") {
		return result, nil
	}

	section := &domain.GitHubCLISection{}

	// Record the config directory if it exists.
	ghConfigDir := filepath.Join(g.homeDir, ".config", "gh")
	if util.DirExists(ghConfigDir) {
		section.ConfigDir = ghConfigDir
	}

	// List installed extensions. Output format per line:
	//   gh-extension-name  <url>  <version>
	// We only take the first whitespace-separated field from each non-empty line.
	lines, err := g.cmd.RunLines(ctx, "gh", "extension", "list")
	if err == nil {
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) > 0 {
				section.Extensions = append(section.Extensions, fields[0])
			}
		}
	}

	result.GitHubCLI = section
	return result, nil
}
