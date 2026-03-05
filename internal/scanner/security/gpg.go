package security

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// GPGScanner scans GPG keys and configuration files.
type GPGScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewGPGScanner creates a new GPGScanner with the given homeDir and CommandRunner.
func NewGPGScanner(homeDir string, cmd util.CommandRunner) *GPGScanner {
	return &GPGScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (g *GPGScanner) Name() string        { return "gpg" }
func (g *GPGScanner) Description() string { return "Scans GPG keys and configuration" }
func (g *GPGScanner) Category() string    { return "security" }

// Scan lists GPG public keys and locates configuration files in ~/.gnupg.
// Returns a ScanResult with the GPG field populated.
func (g *GPGScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: g.Name(),
	}

	if !g.cmd.IsInstalled(ctx, "gpg") {
		return result, nil
	}

	section := &domain.GPGSection{
		Encrypted: true,
	}

	// List public keys and parse "pub" lines for key IDs.
	// Output format (colon-separated): pub:...:...:...<keyid>:...
	// Field index 4 (0-based) holds the key ID.
	output, err := g.cmd.Run(ctx, "gpg", "--list-keys", "--keyid-format", "long", "--with-colons")
	if err == nil && output != "" {
		for _, line := range strings.Split(output, "\n") {
			if !strings.HasPrefix(line, "pub:") {
				continue
			}
			fields := strings.Split(line, ":")
			if len(fields) > 4 && fields[4] != "" {
				section.Keys = append(section.Keys, fields[4])
			}
		}
	}

	// Check for ~/.gnupg/gpg.conf.
	gnupgDir := filepath.Join(g.homeDir, ".gnupg")
	gpgConf := filepath.Join(gnupgDir, "gpg.conf")
	if util.FileExists(gpgConf) {
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:     ".gnupg/gpg.conf",
			BundlePath: "configs/gnupg/gpg.conf",
		})
	}

	// Check for ~/.gnupg/gpg-agent.conf.
	gpgAgentConf := filepath.Join(gnupgDir, "gpg-agent.conf")
	if util.FileExists(gpgAgentConf) {
		section.ConfigFiles = append(section.ConfigFiles, domain.ConfigFile{
			Source:     ".gnupg/gpg-agent.conf",
			BundlePath: "configs/gnupg/gpg-agent.conf",
		})
	}

	result.GPG = section
	return result, nil
}
