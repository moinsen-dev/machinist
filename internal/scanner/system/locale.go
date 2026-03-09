package system

import (
	"context"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// LocaleScanner scans macOS locale, timezone, and hostname settings.
type LocaleScanner struct {
	cmd util.CommandRunner
}

// NewLocaleScanner creates a new LocaleScanner with the given CommandRunner.
func NewLocaleScanner(cmd util.CommandRunner) *LocaleScanner {
	return &LocaleScanner{cmd: cmd}
}

func (s *LocaleScanner) Name() string        { return "locale" }
func (s *LocaleScanner) Description() string  { return "Scans macOS locale, timezone, and hostname settings" }
func (s *LocaleScanner) Category() string     { return "system" }

// Scan reads locale-related settings and returns a ScanResult with the Locale field populated.
func (s *LocaleScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: s.Name(),
	}

	section := &domain.LocaleSection{}

	// Read preferred language
	langOutput, err := s.cmd.Run(ctx, "defaults", "read", "NSGlobalDomain", "AppleLanguages")
	if err == nil {
		section.Language = parseFirstLanguage(langOutput)
	}

	// Read locale/region
	locale, err := s.cmd.Run(ctx, "defaults", "read", "NSGlobalDomain", "AppleLocale")
	if err == nil && locale != "" {
		section.Region = locale
	}

	// Read timezone via /etc/localtime symlink (no sudo required).
	// Falls back to systemsetup if symlink is missing.
	tzOutput, err := s.cmd.Run(ctx, "readlink", "/etc/localtime")
	if err == nil {
		section.Timezone = parseTimezonePath(tzOutput)
	}
	if section.Timezone == "" {
		tzOutput, err = s.cmd.Run(ctx, "systemsetup", "-gettimezone")
		if err == nil && !strings.Contains(tzOutput, "administrator") {
			section.Timezone = parseTimezone(tzOutput)
		}
	}

	// Read computer name
	computerName, err := s.cmd.Run(ctx, "scutil", "--get", "ComputerName")
	if err == nil && computerName != "" {
		section.ComputerName = computerName
	}

	// Read local hostname
	localHostname, err := s.cmd.Run(ctx, "scutil", "--get", "LocalHostName")
	if err == nil && localHostname != "" {
		section.LocalHostname = localHostname
	}

	result.Locale = section
	return result, nil
}

// parseFirstLanguage extracts the first language from the AppleLanguages plist array output.
// Example input:
//
//	(
//	    "en-US"
//	)
func parseFirstLanguage(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// Skip parentheses, empty lines, and bare paren pairs
		if line == "" || line == "(" || line == ")" || line == "()" {
			continue
		}
		// Remove surrounding quotes and trailing comma
		line = strings.Trim(line, "\"',")
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

// parseTimezonePath extracts the timezone from /etc/localtime symlink target.
// Example input: "/var/db/timezone/zoneinfo/Europe/Berlin"
func parseTimezonePath(output string) string {
	output = strings.TrimSpace(output)
	if idx := strings.Index(output, "zoneinfo/"); idx >= 0 {
		return output[idx+len("zoneinfo/"):]
	}
	return ""
}

// parseTimezone extracts the timezone from systemsetup output.
// Example input: "Time Zone: America/Los_Angeles"
func parseTimezone(output string) string {
	if idx := strings.Index(output, ": "); idx >= 0 {
		return strings.TrimSpace(output[idx+2:])
	}
	return strings.TrimSpace(output)
}
