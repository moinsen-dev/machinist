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

// AsdfScanner scans asdf or mise version manager plugins and installed versions.
type AsdfScanner struct {
	cmd     util.CommandRunner
	homeDir string
}

// NewAsdfScanner creates a new AsdfScanner with the given home directory and CommandRunner.
func NewAsdfScanner(homeDir string, cmd util.CommandRunner) *AsdfScanner {
	return &AsdfScanner{homeDir: homeDir, cmd: cmd}
}

func (a *AsdfScanner) Name() string        { return "asdf" }
func (a *AsdfScanner) Description() string { return "Scans asdf/mise plugins and installed versions" }
func (a *AsdfScanner) Category() string    { return "runtimes" }

// Scan checks for asdf or mise and returns a ScanResult with the Asdf field populated.
// asdf is preferred; mise is used as a fallback.
func (a *AsdfScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: a.Name(),
	}

	switch {
	case a.cmd.IsInstalled(ctx, "asdf"):
		section, err := a.scanAsdf(ctx)
		if err != nil {
			return nil, err
		}
		result.Asdf = section
	case a.cmd.IsInstalled(ctx, "mise"):
		section, err := a.scanMise(ctx)
		if err != nil {
			return nil, err
		}
		result.Asdf = section
	default:
		// Neither tool found; return empty result with nil Asdf section.
		return result, nil
	}

	return result, nil
}

// scanAsdf collects plugins and versions using asdf CLI commands.
func (a *AsdfScanner) scanAsdf(ctx context.Context) (*domain.AsdfSection, error) {
	section := &domain.AsdfSection{Manager: "asdf"}

	// Collect plugin names.
	pluginNames, err := a.cmd.RunLines(ctx, "asdf", "plugin", "list")
	if err != nil {
		return nil, err
	}

	for _, name := range pluginNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		// Collect installed versions for each plugin.
		versionLines, err := a.cmd.RunLines(ctx, "asdf", "list", name)
		if err != nil {
			// Skip plugins that fail to list versions; don't abort the whole scan.
			section.Plugins = append(section.Plugins, domain.AsdfPlugin{Name: name})
			continue
		}

		var versions []string
		for _, line := range versionLines {
			v := cleanAsdfVersion(line)
			if v != "" {
				versions = append(versions, v)
			}
		}

		section.Plugins = append(section.Plugins, domain.AsdfPlugin{
			Name:     name,
			Versions: versions,
		})
	}

	// Check for ~/.tool-versions file.
	toolVersionsPath := filepath.Join(a.homeDir, ".tool-versions")
	if _, err := os.Stat(toolVersionsPath); err == nil {
		section.ToolVersionsFile = toolVersionsPath
	}

	return section, nil
}

// scanMise collects plugins and versions using mise CLI commands.
func (a *AsdfScanner) scanMise(ctx context.Context) (*domain.AsdfSection, error) {
	section := &domain.AsdfSection{Manager: "mise"}

	// Collect plugin names.
	pluginNames, err := a.cmd.RunLines(ctx, "mise", "plugins", "list")
	if err != nil {
		return nil, err
	}

	// Collect all installed versions in one call.
	// Output format: "tool  version  source" (whitespace-separated fields).
	versionLines, err := a.cmd.RunLines(ctx, "mise", "list")
	if err != nil {
		return nil, err
	}

	// Build a map of plugin -> []version from `mise list` output.
	pluginVersions := parseMiseList(versionLines)

	for _, name := range pluginNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		section.Plugins = append(section.Plugins, domain.AsdfPlugin{
			Name:     name,
			Versions: pluginVersions[name],
		})
	}

	// Check for ~/.tool-versions file.
	toolVersionsPath := filepath.Join(a.homeDir, ".tool-versions")
	if _, err := os.Stat(toolVersionsPath); err == nil {
		section.ToolVersionsFile = toolVersionsPath
	}

	return section, nil
}

// cleanAsdfVersion strips leading whitespace and the '*' current-version marker
// from a single line of `asdf list <plugin>` output.
func cleanAsdfVersion(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimSpace(line)
	return line
}

// parseMiseList parses the output of `mise list` and returns a map of
// plugin name -> []version. Each non-empty line has the format:
//
//	tool  version  [source]
func parseMiseList(lines []string) map[string][]string {
	result := make(map[string][]string)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		tool := fields[0]
		version := fields[1]
		result[tool] = append(result[tool], version)
	}
	return result
}
