package runtimes

import (
	"context"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// RustScanner scans Rust toolchains, components, and cargo-installed packages.
type RustScanner struct {
	cmd util.CommandRunner
}

// NewRustScanner creates a new RustScanner with the given CommandRunner.
func NewRustScanner(cmd util.CommandRunner) *RustScanner {
	return &RustScanner{cmd: cmd}
}

func (r *RustScanner) Name() string        { return "rust" }
func (r *RustScanner) Description() string  { return "Scans Rust toolchains and cargo packages" }
func (r *RustScanner) Category() string     { return "runtimes" }

// Scan runs rustup and cargo commands and returns a ScanResult with the Rust field populated.
func (r *RustScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: r.Name(),
	}

	if !r.cmd.IsInstalled(ctx, "rustup") {
		return result, nil
	}

	section := &domain.RustSection{}

	// Toolchains
	toolchainLines, err := r.cmd.RunLines(ctx, "rustup", "toolchain", "list")
	if err != nil {
		return nil, err
	}
	for _, line := range toolchainLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, "(default)") {
			name := strings.TrimSpace(strings.TrimSuffix(line, "(default)"))
			section.DefaultToolchain = name
			section.Toolchains = append(section.Toolchains, name)
		} else {
			section.Toolchains = append(section.Toolchains, line)
		}
	}

	// Components
	components, err := r.cmd.RunLines(ctx, "rustup", "component", "list", "--installed")
	if err != nil {
		return nil, err
	}
	section.Components = components

	// Cargo packages
	cargoOutput, err := r.cmd.Run(ctx, "cargo", "install", "--list")
	if err != nil {
		return nil, err
	}
	section.CargoPackages = parseCargoInstallList(cargoOutput)

	result.Rust = section
	return result, nil
}

// parseCargoInstallList parses the output of `cargo install --list`.
// Lines matching "name vX.Y.Z:" are treated as package entries.
// Indented lines (binaries) are skipped.
func parseCargoInstallList(output string) []domain.Package {
	var pkgs []domain.Package
	if output == "" {
		return pkgs
	}
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.HasSuffix(line, ":") {
			continue
		}
		// Strip trailing colon
		line = strings.TrimSuffix(line, ":")
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		pkgs = append(pkgs, domain.Package{
			Name:    parts[0],
			Version: parts[1],
		})
	}
	return pkgs
}
