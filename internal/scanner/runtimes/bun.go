package runtimes

import (
	"context"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// BunScanner scans Bun version and globally installed packages.
type BunScanner struct {
	cmd util.CommandRunner
}

// NewBunScanner creates a new BunScanner with the given CommandRunner.
func NewBunScanner(cmd util.CommandRunner) *BunScanner {
	return &BunScanner{cmd: cmd}
}

func (b *BunScanner) Name() string        { return "bun" }
func (b *BunScanner) Description() string  { return "Scans Bun version and globally installed packages" }
func (b *BunScanner) Category() string     { return "runtimes" }

// Scan detects the Bun version and lists globally installed packages.
func (b *BunScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: b.Name(),
	}

	if !b.cmd.IsInstalled(ctx, "bun") {
		return result, nil
	}

	section := &domain.BunSection{}

	// Version: `bun --version` outputs just the version number, e.g. "1.0.25"
	version, err := b.cmd.Run(ctx, "bun", "--version")
	if err != nil {
		return nil, err
	}
	section.Version = strings.TrimSpace(version)

	// Global packages: `bun pm ls -g`
	// Output format (tree-style):
	//   /path/to/global/node_modules (global)
	//   ├── package-name@1.2.3
	//   └── another-pkg@0.5.0
	lines, err := b.cmd.RunLines(ctx, "bun", "pm", "ls", "-g")
	if err == nil {
		section.GlobalPackages = parseBunGlobalList(lines)
	}
	// A non-zero exit from `bun pm ls -g` is not fatal — treat as empty list.

	result.Bun = section
	return result, nil
}

// parseBunGlobalList parses the output lines of `bun pm ls -g`.
// It handles tree-style output with "├──" / "└──" prefixes as well as
// plain "  package@version" lines. Each entry produces a Package with
// Name and Version separated at the last "@" that is not the first
// character (scoped packages start with "@").
func parseBunGlobalList(lines []string) []domain.Package {
	var pkgs []domain.Package
	for _, line := range lines {
		// Strip tree drawing characters and surrounding whitespace.
		cleaned := strings.TrimLeft(line, " \t")
		cleaned = strings.TrimPrefix(cleaned, "├── ")
		cleaned = strings.TrimPrefix(cleaned, "└── ")
		cleaned = strings.TrimPrefix(cleaned, "├─ ")
		cleaned = strings.TrimPrefix(cleaned, "└─ ")
		cleaned = strings.TrimSpace(cleaned)

		if cleaned == "" {
			continue
		}

		// Skip header lines like "/path/to/node_modules (global)"
		if strings.HasPrefix(cleaned, "/") || strings.Contains(cleaned, "(global)") {
			continue
		}

		// Skip lines that don't look like a package entry.
		// A valid entry contains at least one non-space character.
		if strings.ContainsAny(cleaned, " \t") && !strings.Contains(cleaned, "@") {
			continue
		}

		name, version := splitBunPackageEntry(cleaned)
		if name == "" {
			continue
		}
		pkgs = append(pkgs, domain.Package{
			Name:    name,
			Version: version,
		})
	}
	return pkgs
}

// splitBunPackageEntry splits a "name@version" entry into name and version.
// Scoped packages (e.g. "@scope/pkg@1.0.0") start with "@"; the version
// separator is the last "@" after the first character.
func splitBunPackageEntry(entry string) (name, version string) {
	// Find the last "@" that is not at position 0 (to handle scoped packages).
	idx := strings.LastIndex(entry, "@")
	if idx <= 0 {
		// No version separator found; treat whole string as name.
		return strings.TrimSpace(entry), ""
	}
	return strings.TrimSpace(entry[:idx]), strings.TrimSpace(entry[idx+1:])
}
