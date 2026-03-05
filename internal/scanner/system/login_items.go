package system

import (
	"context"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// LoginItemsScanner scans macOS login items (apps that start at login).
type LoginItemsScanner struct {
	cmd util.CommandRunner
}

// NewLoginItemsScanner creates a new LoginItemsScanner with the given CommandRunner.
func NewLoginItemsScanner(cmd util.CommandRunner) *LoginItemsScanner {
	return &LoginItemsScanner{cmd: cmd}
}

func (s *LoginItemsScanner) Name() string        { return "login-items" }
func (s *LoginItemsScanner) Description() string  { return "Scans macOS login items" }
func (s *LoginItemsScanner) Category() string     { return "system" }

// Scan reads login items via AppleScript and returns a ScanResult with the LoginItems field populated.
func (s *LoginItemsScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	section := &domain.LoginItemsSection{}

	output, err := s.cmd.Run(ctx, "osascript", "-e", `tell application "System Events" to get name of every login item`)
	if err != nil {
		// TCC or other permission error — return empty section (scanner ran but couldn't read)
		result.LoginItems = section
		return result, nil
	}

	if output != "" {
		apps := parseLoginItems(output)
		section.Apps = apps
	}

	result.LoginItems = section
	return result, nil
}

// parseLoginItems parses the comma-separated list of app names from osascript output.
// Example input: "Dropbox, Alfred 5, Docker"
func parseLoginItems(output string) []string {
	parts := strings.Split(output, ",")
	var apps []string
	for _, p := range parts {
		name := strings.TrimSpace(p)
		if name != "" {
			apps = append(apps, name)
		}
	}
	return apps
}
