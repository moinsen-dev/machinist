package runtimes

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// PythonScanner scans Python versions and globally installed packages.
type PythonScanner struct {
	cmd util.CommandRunner
}

// NewPythonScanner creates a new PythonScanner with the given CommandRunner.
func NewPythonScanner(cmd util.CommandRunner) *PythonScanner {
	return &PythonScanner{cmd: cmd}
}

func (p *PythonScanner) Name() string        { return "python" }
func (p *PythonScanner) Description() string  { return "Scans Python versions and global packages" }
func (p *PythonScanner) Category() string     { return "runtimes" }

// pipPackage is used for JSON deserialization of pip list output.
type pipPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Scan detects the Python version manager, installed versions, default version,
// and globally installed pip packages.
func (p *PythonScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: p.Name(),
	}

	section := &domain.PythonSection{}

	// 1. Detect manager and gather versions
	switch {
	case p.cmd.IsInstalled(ctx, "pyenv"):
		section.Manager = "pyenv"
		p.scanPyenv(ctx, section)
	case p.cmd.IsInstalled(ctx, "uv"):
		section.Manager = "uv"
		p.scanUv(ctx, section)
	default:
		p.scanSystem(ctx, section)
	}

	// If no versions and no default version found, Python is not available
	if len(section.Versions) == 0 && section.DefaultVersion == "" {
		return result, nil
	}

	// 2. Get global packages via pip
	p.scanGlobalPackages(ctx, section)

	result.Python = section
	return result, nil
}

// scanPyenv gathers versions and default version from pyenv.
func (p *PythonScanner) scanPyenv(ctx context.Context, section *domain.PythonSection) {
	versions, err := p.cmd.RunLines(ctx, "pyenv", "versions", "--bare")
	if err == nil {
		section.Versions = versions
	}

	defaultVer, err := p.cmd.Run(ctx, "pyenv", "global")
	if err == nil && defaultVer != "" {
		section.DefaultVersion = defaultVer
	}
}

// scanUv gathers versions from uv python list.
func (p *PythonScanner) scanUv(ctx context.Context, section *domain.PythonSection) {
	lines, err := p.cmd.RunLines(ctx, "uv", "python", "list", "--only-installed")
	if err == nil {
		for _, line := range lines {
			ver := extractUvVersion(line)
			if ver != "" {
				section.Versions = append(section.Versions, ver)
			}
		}
	}
}

// extractUvVersion extracts a version string from uv python list output.
// Each line looks like: "cpython-3.12.1-macos-aarch64-none    /path/to/python3"
func extractUvVersion(line string) string {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return ""
	}
	// The first field is like "cpython-3.12.1-macos-aarch64-none"
	parts := strings.SplitN(fields[0], "-", 3)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// scanSystem tries to detect a system Python installation.
func (p *PythonScanner) scanSystem(ctx context.Context, section *domain.PythonSection) {
	out, err := p.cmd.Run(ctx, "python3", "--version")
	if err != nil {
		return
	}
	// Output is like "Python 3.12.1"
	ver := strings.TrimPrefix(out, "Python ")
	if ver != out && ver != "" {
		section.DefaultVersion = ver
	}
}

// scanGlobalPackages parses pip list --format=json output.
func (p *PythonScanner) scanGlobalPackages(ctx context.Context, section *domain.PythonSection) {
	out, err := p.cmd.Run(ctx, "pip", "list", "--format=json")
	if err != nil {
		return
	}

	var pkgs []pipPackage
	if err := json.Unmarshal([]byte(out), &pkgs); err != nil {
		return
	}

	for _, pkg := range pkgs {
		section.GlobalPackages = append(section.GlobalPackages, domain.Package{
			Name:    pkg.Name,
			Version: pkg.Version,
		})
	}
}
