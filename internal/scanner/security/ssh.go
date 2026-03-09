package security

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// SSHScanner scans SSH keys and configuration from the user's ~/.ssh directory.
type SSHScanner struct {
	homeDir string
}

// NewSSHScanner creates a new SSHScanner that scans the given homeDir.
func NewSSHScanner(homeDir string) *SSHScanner {
	return &SSHScanner{homeDir: homeDir}
}

func (s *SSHScanner) Name() string        { return "ssh" }
func (s *SSHScanner) Description() string { return "Scans SSH keys and configuration" }
func (s *SSHScanner) Category() string    { return "security" }

// Scan lists SSH key files and configuration in ~/.ssh and returns a ScanResult
// with the SSH field populated.
func (s *SSHScanner) Scan(_ context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	sshDir := filepath.Join(s.homeDir, ".ssh")
	if !util.DirExists(sshDir) {
		return result, nil
	}

	section := &domain.SSHSection{
		Encrypted: true,
	}

	// Check for ~/.ssh/config — store as relative path for portability.
	configPath := filepath.Join(sshDir, "config")
	if util.FileExists(configPath) {
		section.ConfigFile = filepath.Join(".ssh", "config")
	}

	// Check for ~/.ssh/known_hosts.
	knownHostsPath := filepath.Join(sshDir, "known_hosts")
	if util.FileExists(knownHostsPath) {
		section.KnownHosts = filepath.Join(".ssh", "known_hosts")
	}

	// List key files: names matching id_* but not ending in .pub.
	// Store as bare filenames (e.g. "id_ed25519") — the bundler resolves
	// them relative to ~/.ssh/.
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		// Directory exists but cannot be read — still return what we have.
		result.SSH = section
		return result, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "id_") && !strings.HasSuffix(name, ".pub") {
			section.Keys = append(section.Keys, name)
		}
	}

	result.SSH = section
	return result, nil
}
