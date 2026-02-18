package system

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// ScheduledScanner scans crontab entries and user-level LaunchAgents.
type ScheduledScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewScheduledScanner creates a new ScheduledScanner.
func NewScheduledScanner(homeDir string, cmd util.CommandRunner) *ScheduledScanner {
	return &ScheduledScanner{
		homeDir: homeDir,
		cmd:     cmd,
	}
}

func (s *ScheduledScanner) Name() string        { return "scheduled" }
func (s *ScheduledScanner) Description() string { return "Scans crontab entries and LaunchAgents" }
func (s *ScheduledScanner) Category() string    { return "system" }

// Scan collects crontab entries and LaunchAgent plist files.
func (s *ScheduledScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{ScannerName: s.Name()}

	// 1. Scan crontab entries.
	crontabSection := s.scanCrontab(ctx)
	if len(crontabSection.Entries) > 0 {
		result.Crontab = crontabSection
	}

	// 2. Scan LaunchAgents.
	launchAgentsSection := s.scanLaunchAgents()
	if len(launchAgentsSection.Plists) > 0 {
		result.LaunchAgents = launchAgentsSection
	}

	return result, nil
}

// scanCrontab runs `crontab -l` and parses non-empty, non-comment lines.
func (s *ScheduledScanner) scanCrontab(ctx context.Context) *domain.CrontabSection {
	section := &domain.CrontabSection{}

	output, err := s.cmd.Run(ctx, "crontab", "-l")
	if err != nil {
		return section
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		section.Entries = append(section.Entries, line)
	}

	return section
}

// scanLaunchAgents lists .plist files in ~/Library/LaunchAgents/.
func (s *ScheduledScanner) scanLaunchAgents() *domain.LaunchAgentsSection {
	section := &domain.LaunchAgentsSection{}

	launchAgentsDir := filepath.Join(s.homeDir, "Library", "LaunchAgents")
	entries, err := os.ReadDir(launchAgentsDir)
	if err != nil {
		return section
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".plist") {
			continue
		}
		absPath := filepath.Join(launchAgentsDir, entry.Name())
		source := filepath.Join("Library", "LaunchAgents", entry.Name())

		hash, _ := util.ContentHash(absPath)
		section.Plists = append(section.Plists, domain.ConfigFile{
			Source:      source,
			BundlePath:  filepath.Join("configs", "launchagents", entry.Name()),
			ContentHash: hash,
		})
	}

	return section
}
