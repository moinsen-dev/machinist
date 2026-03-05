package runtimes

import (
	"context"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// FlutterScanner scans Flutter SDK channel, version, and Dart global packages.
type FlutterScanner struct {
	cmd util.CommandRunner
}

// NewFlutterScanner creates a new FlutterScanner with the given CommandRunner.
func NewFlutterScanner(cmd util.CommandRunner) *FlutterScanner {
	return &FlutterScanner{cmd: cmd}
}

func (f *FlutterScanner) Name() string        { return "flutter" }
func (f *FlutterScanner) Description() string { return "Scans Flutter SDK and Dart global packages" }
func (f *FlutterScanner) Category() string    { return "runtimes" }

// Scan runs flutter --version and dart pub global list to populate FlutterSection.
func (f *FlutterScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: f.Name(),
	}

	if !f.cmd.IsInstalled(ctx, "flutter") {
		return result, nil
	}

	section := &domain.FlutterSection{}

	// Get Flutter version and channel.
	versionOutput, err := f.cmd.Run(ctx, "flutter", "--version")
	if err != nil {
		return nil, err
	}
	section.Version, section.Channel = parseFlutterVersion(versionOutput)

	// List Dart global packages.
	dartPkgLines, err := f.cmd.RunLines(ctx, "dart", "pub", "global", "list")
	if err == nil {
		for _, line := range dartPkgLines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Lines are in format: "package_name X.Y.Z"
			// We only need the package name.
			fields := strings.Fields(line)
			if len(fields) > 0 {
				section.DartGlobalPackages = append(section.DartGlobalPackages, fields[0])
			}
		}
	}

	result.Flutter = section
	return result, nil
}

// parseFlutterVersion extracts the version and channel from `flutter --version` output.
// Example first line: "Flutter 3.19.1 • channel stable • https://github.com/flutter/flutter.git"
// Returns (version, channel).
func parseFlutterVersion(output string) (version, channel string) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Flutter ") {
			continue
		}
		// Split on " • " to get the parts.
		parts := strings.Split(line, " • ")
		if len(parts) == 0 {
			continue
		}

		// First part: "Flutter X.Y.Z"
		flutterPart := strings.Fields(parts[0])
		if len(flutterPart) >= 2 {
			version = flutterPart[1]
		}

		// Second part: "channel <name>"
		if len(parts) >= 2 {
			channelPart := strings.Fields(parts[1])
			if len(channelPart) >= 2 && channelPart[0] == "channel" {
				channel = channelPart[1]
			}
		}

		return version, channel
	}
	return "", ""
}
