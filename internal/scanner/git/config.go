package git

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// GitConfigScanner scans git global configuration and settings.
type GitConfigScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewGitConfigScanner creates a new GitConfigScanner with the given homeDir and CommandRunner.
func NewGitConfigScanner(homeDir string, cmd util.CommandRunner) *GitConfigScanner {
	return &GitConfigScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (g *GitConfigScanner) Name() string        { return "git-config" }
func (g *GitConfigScanner) Description() string { return "Scans git configuration and settings" }
func (g *GitConfigScanner) Category() string    { return "git" }

// Scan reads git global configuration and returns a ScanResult with the Git field populated.
func (g *GitConfigScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: g.Name(),
	}

	if !g.cmd.IsInstalled(ctx, "git") {
		return result, nil
	}

	section := &domain.GitSection{}

	// Check for ~/.gitconfig and add as a ConfigFile if it exists.
	gitconfigPath := filepath.Join(g.homeDir, ".gitconfig")
	if util.FileExists(gitconfigPath) {
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:     ".gitconfig",
			BundlePath: "configs/.gitconfig",
		})
	}

	// Check for a global excludesfile and add it as a ConfigFile if non-empty.
	excludesFile, err := g.cmd.Run(ctx, "git", "config", "--global", "core.excludesfile")
	if err == nil && excludesFile != "" {
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:     excludesFile,
			BundlePath: "configs/git-excludes",
		})
	}

	// Determine signing method from user.signingkey and gpg.format.
	signingKey, err := g.cmd.Run(ctx, "git", "config", "--global", "user.signingkey")
	if err == nil && signingKey != "" {
		// Default to "gpg" unless gpg.format overrides it.
		section.SigningMethod = "gpg"
	}

	gpgFormat, err := g.cmd.Run(ctx, "git", "config", "--global", "gpg.format")
	if err == nil && gpgFormat != "" {
		format := strings.TrimSpace(gpgFormat)
		if format == "ssh" || format == "gpg" {
			section.SigningMethod = format
		}
	}

	// Credential helper.
	credHelper, err := g.cmd.Run(ctx, "git", "config", "--global", "credential.helper")
	if err == nil && credHelper != "" {
		section.CredentialHelper = credHelper
	}

	// Template directory.
	templateDir, err := g.cmd.Run(ctx, "git", "config", "--global", "init.templateDir")
	if err == nil && templateDir != "" {
		section.TemplateDir = templateDir
	}

	// Only set the section if we found anything meaningful.
	if len(section.ConfigFiles) > 0 ||
		section.SigningMethod != "" ||
		section.CredentialHelper != "" ||
		section.TemplateDir != "" {
		result.Git = section
	}

	return result, nil
}
