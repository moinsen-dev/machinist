package shell

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// tpmPluginRe matches TPM plugin declarations such as:
//
//	set -g @plugin 'tmux-plugins/tpm'
//	set -g @plugin "tmux-plugins/sensible"
var tpmPluginRe = regexp.MustCompile(`set\s+-g\s+@plugin\s+['"]([^'"]+)['"]`)

// TmuxScanner scans tmux configuration files and TPM plugins.
type TmuxScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewTmuxScanner creates a new TmuxScanner.
func NewTmuxScanner(homeDir string, cmd util.CommandRunner) *TmuxScanner {
	return &TmuxScanner{homeDir: homeDir, cmd: cmd}
}

func (s *TmuxScanner) Name() string        { return "tmux" }
func (s *TmuxScanner) Description() string  { return "Scans tmux configuration and plugins" }
func (s *TmuxScanner) Category() string     { return "shell" }

// Scan collects tmux config files and extracts TPM plugin names from them.
// Returns an empty result (nil Tmux section) if tmux is not installed.
func (s *TmuxScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	if !s.cmd.IsInstalled(ctx, "tmux") {
		return result, nil
	}

	section := &domain.TmuxSection{}

	// Candidate config file locations, in order.
	configCandidates := []struct {
		rel string // relative to homeDir
	}{
		{rel: ".tmux.conf"},
		{rel: ".config/tmux/tmux.conf"},
	}

	var combinedContent strings.Builder

	for _, c := range configCandidates {
		absPath := filepath.Join(s.homeDir, c.rel)
		if !util.FileExists(absPath) {
			continue
		}

		hash, err := util.ContentHash(absPath)
		cf := domain.ConfigFile{
			Source:     c.rel,
			BundlePath: filepath.Join("configs", "tmux", filepath.Base(c.rel)),
		}
		if err == nil {
			cf.ContentHash = hash
		}
		section.ConfigFiles = append(section.ConfigFiles, cf)

		// Read content for plugin parsing.
		data, err := os.ReadFile(absPath)
		if err == nil {
			combinedContent.Write(data)
			combinedContent.WriteByte('\n')
		}
	}

	// Extract TPM plugin names from the combined config content.
	section.TPMPlugins = parseTPMPlugins(combinedContent.String())

	// Only attach a section if tmux is installed (we already checked), even when
	// no config files are found – this signals "tmux present, no config".
	result.Tmux = section
	return result, nil
}

// parseTPMPlugins scans config content for `set -g @plugin '...'` lines and
// returns the extracted plugin names.
func parseTPMPlugins(content string) []string {
	var plugins []string
	seen := make(map[string]bool)

	for _, match := range tpmPluginRe.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		name := strings.TrimSpace(match[1])
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		plugins = append(plugins, name)
	}
	return plugins
}
