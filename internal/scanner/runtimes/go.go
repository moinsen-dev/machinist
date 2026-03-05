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

// GoScanner scans Go version and globally installed tools.
type GoScanner struct {
	cmd util.CommandRunner
}

// NewGoScanner creates a new GoScanner with the given CommandRunner.
func NewGoScanner(cmd util.CommandRunner) *GoScanner {
	return &GoScanner{cmd: cmd}
}

func (g *GoScanner) Name() string        { return "go" }
func (g *GoScanner) Description() string { return "Scans Go version and globally installed tools" }
func (g *GoScanner) Category() string    { return "runtimes" }

// Scan runs go version and go env GOPATH, then lists $GOPATH/bin/ for global tools.
func (g *GoScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: g.Name(),
	}

	if !g.cmd.IsInstalled(ctx, "go") {
		return result, nil
	}

	section := &domain.GoSection{}

	// Parse Go version.
	versionOutput, err := g.cmd.Run(ctx, "go", "version")
	if err != nil {
		return nil, err
	}
	section.Version = parseGoVersion(versionOutput)

	// Get GOPATH.
	gopath, err := g.cmd.Run(ctx, "go", "env", "GOPATH")
	if err != nil {
		return nil, err
	}
	gopath = strings.TrimSpace(gopath)

	// List binaries in $GOPATH/bin/.
	if gopath != "" {
		binDir := filepath.Join(gopath, "bin")
		entries, err := os.ReadDir(binDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				section.GlobalPackages = append(section.GlobalPackages, domain.Package{
					Name: entry.Name(),
				})
			}
		}
	}

	result.GoLang = section
	return result, nil
}

// parseGoVersion extracts the version number from `go version` output.
// Example: "go version go1.24.1 darwin/arm64" -> "1.24.1"
func parseGoVersion(output string) string {
	// Output format: "go version goX.Y.Z OS/ARCH"
	fields := strings.Fields(output)
	for _, f := range fields {
		if strings.HasPrefix(f, "go") && len(f) > 2 {
			ver := strings.TrimPrefix(f, "go")
			// Only return if it looks like a version (starts with a digit).
			if len(ver) > 0 && ver[0] >= '0' && ver[0] <= '9' {
				return ver
			}
		}
	}
	return output
}
