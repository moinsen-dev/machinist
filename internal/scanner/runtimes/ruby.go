package runtimes

import (
	"context"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// RubyScanner scans Ruby version manager, installed versions, and globally installed gems.
type RubyScanner struct {
	cmd util.CommandRunner
}

// NewRubyScanner creates a new RubyScanner with the given CommandRunner.
func NewRubyScanner(cmd util.CommandRunner) *RubyScanner {
	return &RubyScanner{cmd: cmd}
}

func (r *RubyScanner) Name() string        { return "ruby" }
func (r *RubyScanner) Description() string  { return "Scans Ruby version manager, versions, and global gems" }
func (r *RubyScanner) Category() string     { return "runtimes" }

// Scan detects the Ruby version manager (rbenv, rvm, or system), installed
// versions, the default/current version, and globally installed gems.
func (r *RubyScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: r.Name(),
	}

	section := &domain.RubySection{}

	switch {
	case r.cmd.IsInstalled(ctx, "rbenv"):
		section.Manager = "rbenv"
		r.scanRbenv(ctx, section)

	case r.cmd.IsInstalled(ctx, "rvm"):
		section.Manager = "rvm"
		r.scanRvm(ctx, section)

	default:
		// Fall back to system ruby.
		ver, err := r.cmd.Run(ctx, "ruby", "--version")
		if err != nil {
			// No ruby at all.
			return result, nil
		}
		// Output: "ruby 3.2.2 (2023-03-30 revision e51014f9c0) [arm64-darwin22]"
		// Extract just the version token (second word).
		fields := strings.Fields(ver)
		if len(fields) >= 2 {
			section.DefaultVersion = fields[1]
			section.Versions = []string{fields[1]}
		} else {
			section.DefaultVersion = ver
			section.Versions = []string{ver}
		}
	}

	// Collect globally installed gems regardless of manager.
	gems, err := r.scanGems(ctx)
	if err != nil {
		return nil, err
	}
	section.GlobalGems = gems

	result.Ruby = section
	return result, nil
}

// scanRbenv populates the section using rbenv commands.
func (r *RubyScanner) scanRbenv(ctx context.Context, section *domain.RubySection) {
	// `rbenv versions --bare` → one version per line
	versions, err := r.cmd.RunLines(ctx, "rbenv", "versions", "--bare")
	if err == nil {
		for _, v := range versions {
			v = strings.TrimSpace(v)
			if v != "" {
				section.Versions = append(section.Versions, v)
			}
		}
	}

	// `rbenv global` → default version
	def, err := r.cmd.Run(ctx, "rbenv", "global")
	if err == nil {
		section.DefaultVersion = strings.TrimSpace(def)
	}
}

// scanRvm populates the section using rvm commands.
func (r *RubyScanner) scanRvm(ctx context.Context, section *domain.RubySection) {
	// `rvm list strings` → one version per line
	versions, err := r.cmd.RunLines(ctx, "rvm", "list", "strings")
	if err == nil {
		for _, v := range versions {
			v = strings.TrimSpace(v)
			if v != "" {
				section.Versions = append(section.Versions, v)
			}
		}
	}

	// `rvm current` → default/current version
	def, err := r.cmd.Run(ctx, "rvm", "current")
	if err == nil {
		section.DefaultVersion = strings.TrimSpace(def)
	}
}

// scanGems returns globally installed gems from `gem list --no-versions`.
func (r *RubyScanner) scanGems(ctx context.Context) ([]domain.Package, error) {
	lines, err := r.cmd.RunLines(ctx, "gem", "list", "--no-versions")
	if err != nil {
		// gem not available is not a fatal error.
		return nil, nil
	}

	var pkgs []domain.Package
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// `gem list --no-versions` can prefix lines with "*** LOCAL GEMS ***" header.
		if strings.HasPrefix(line, "***") {
			continue
		}
		pkgs = append(pkgs, domain.Package{Name: line})
	}
	return pkgs, nil
}
