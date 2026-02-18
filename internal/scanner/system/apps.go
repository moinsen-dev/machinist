package system

import (
	"context"
	"strconv"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// AppsScanner scans Mac App Store apps via the mas CLI tool.
type AppsScanner struct {
	cmd util.CommandRunner
}

// NewAppsScanner creates a new AppsScanner.
func NewAppsScanner(cmd util.CommandRunner) *AppsScanner {
	return &AppsScanner{cmd: cmd}
}

func (s *AppsScanner) Name() string        { return "apps" }
func (s *AppsScanner) Description() string { return "Scans Mac App Store apps via mas" }
func (s *AppsScanner) Category() string    { return "system" }

// Scan checks for mas CLI and lists installed App Store apps.
func (s *AppsScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	if !s.cmd.IsInstalled(ctx, "mas") {
		return &scanner.ScanResult{ScannerName: s.Name()}, nil
	}

	lines, err := s.cmd.RunLines(ctx, "mas", "list")
	if err != nil {
		return &scanner.ScanResult{ScannerName: s.Name()}, nil
	}

	var apps []domain.InstalledApp
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		app, ok := parseMasLine(line)
		if !ok {
			continue
		}
		apps = append(apps, app)
	}

	if len(apps) == 0 {
		return &scanner.ScanResult{ScannerName: s.Name()}, nil
	}

	return &scanner.ScanResult{
		ScannerName: s.Name(),
		Apps: &domain.AppsSection{
			AppStore: apps,
		},
	}, nil
}

// parseMasLine parses a single line from `mas list` output.
// Format: "497799835  Xcode (15.2)"
// Returns the InstalledApp and true if parsing succeeded.
func parseMasLine(line string) (domain.InstalledApp, bool) {
	// Find the first space to separate ID from the rest.
	idxSpace := strings.IndexByte(line, ' ')
	if idxSpace < 0 {
		return domain.InstalledApp{}, false
	}

	idStr := line[:idxSpace]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return domain.InstalledApp{}, false
	}

	// The rest after the ID and whitespace is "Name (version)"
	rest := strings.TrimLeft(line[idxSpace:], " ")
	if rest == "" {
		return domain.InstalledApp{}, false
	}

	// Find the last " (" to separate name from version.
	// This handles names that contain parentheses, e.g. "Some App (Pro) (2.1.0)".
	lastParen := strings.LastIndex(rest, " (")
	if lastParen < 0 {
		return domain.InstalledApp{}, false
	}

	name := rest[:lastParen]
	if name == "" {
		return domain.InstalledApp{}, false
	}

	return domain.InstalledApp{
		ID:     id,
		Name:   name,
		Source: "mas",
	}, true
}
