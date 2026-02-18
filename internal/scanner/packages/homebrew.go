package packages

import (
	"context"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// HomebrewScanner scans Homebrew packages, casks, taps, and services.
type HomebrewScanner struct {
	cmd util.CommandRunner
}

// NewHomebrewScanner creates a new HomebrewScanner with the given CommandRunner.
func NewHomebrewScanner(cmd util.CommandRunner) *HomebrewScanner {
	return &HomebrewScanner{cmd: cmd}
}

func (h *HomebrewScanner) Name() string        { return "homebrew" }
func (h *HomebrewScanner) Description() string  { return "Scans Homebrew packages, casks, taps, and services" }
func (h *HomebrewScanner) Category() string     { return "packages" }

// Scan runs brew commands and returns a ScanResult with the Homebrew field populated.
func (h *HomebrewScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: h.Name(),
	}

	if !h.cmd.IsInstalled(ctx, "brew") {
		return result, nil
	}

	section := &domain.HomebrewSection{}

	// Formulae
	formulae, err := h.cmd.RunLines(ctx, "brew", "list", "--formula", "--versions")
	if err != nil {
		return nil, err
	}
	for _, line := range formulae {
		parts := strings.SplitN(line, " ", 2)
		pkg := domain.Package{Name: parts[0]}
		if len(parts) > 1 {
			pkg.Version = parts[1]
		}
		section.Formulae = append(section.Formulae, pkg)
	}

	// Casks
	casks, err := h.cmd.RunLines(ctx, "brew", "list", "--cask")
	if err != nil {
		return nil, err
	}
	for _, line := range casks {
		section.Casks = append(section.Casks, domain.Package{Name: line})
	}

	// Taps
	taps, err := h.cmd.RunLines(ctx, "brew", "tap")
	if err != nil {
		return nil, err
	}
	section.Taps = taps

	// Services
	lines, err := h.cmd.RunLines(ctx, "brew", "services", "list")
	if err != nil {
		return nil, err
	}
	for i, line := range lines {
		// Skip header line
		if i == 0 && strings.HasPrefix(line, "Name") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		entry := domain.ServiceEntry{Name: fields[0]}
		if len(fields) > 1 {
			entry.Status = fields[1]
		}
		section.Services = append(section.Services, entry)
	}

	result.Homebrew = section
	return result, nil
}
