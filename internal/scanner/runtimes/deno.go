package runtimes

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// DenoScanner scans Deno version and globally installed scripts.
type DenoScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewDenoScanner creates a new DenoScanner with the given home directory and CommandRunner.
func NewDenoScanner(homeDir string, cmd util.CommandRunner) *DenoScanner {
	return &DenoScanner{homeDir: homeDir, cmd: cmd}
}

func (d *DenoScanner) Name() string        { return "deno" }
func (d *DenoScanner) Description() string  { return "Scans Deno version and globally installed scripts" }
func (d *DenoScanner) Category() string     { return "runtimes" }

// Scan detects the Deno version and lists globally installed scripts from ~/.deno/bin/.
func (d *DenoScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: d.Name(),
	}

	if !d.cmd.IsInstalled(ctx, "deno") {
		return result, nil
	}

	section := &domain.DenoSection{}

	// Version: `deno --version` outputs something like:
	//   deno 1.40.2
	//   v8 12.1.285.27
	//   typescript 5.3.3
	versionOutput, err := d.cmd.Run(ctx, "deno", "--version")
	if err != nil {
		return nil, err
	}
	if versionOutput != "" {
		firstLine := strings.SplitN(versionOutput, "\n", 2)[0]
		firstLine = strings.TrimSpace(firstLine)
		// Strip leading "deno " prefix.
		if strings.HasPrefix(firstLine, "deno ") {
			section.Version = strings.TrimPrefix(firstLine, "deno ")
		} else {
			section.Version = firstLine
		}
	}

	// Global scripts: read files in ~/.deno/bin/
	binDir := filepath.Join(d.homeDir, ".deno", "bin")
	entries, err := os.ReadDir(binDir)
	if err != nil {
		// Directory doesn't exist or can't be read — not a hard error.
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			section.GlobalPackages = append(section.GlobalPackages, domain.Package{
				Name: entry.Name(),
			})
		}
	}

	result.Deno = section
	return result, nil
}
