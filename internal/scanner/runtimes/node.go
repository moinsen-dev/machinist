package runtimes

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
)

// NodeScanner scans Node.js versions and global packages.
type NodeScanner struct {
	homeDir string
	cmd     util.CommandRunner
}

// NewNodeScanner creates a new NodeScanner with the given home directory and CommandRunner.
func NewNodeScanner(homeDir string, cmd util.CommandRunner) *NodeScanner {
	return &NodeScanner{homeDir: homeDir, cmd: cmd}
}

func (n *NodeScanner) Name() string        { return "node" }
func (n *NodeScanner) Description() string  { return "Scans Node.js versions and global packages" }
func (n *NodeScanner) Category() string     { return "runtimes" }

// Scan detects the Node.js version manager, installed versions, default version,
// and globally installed npm packages.
func (n *NodeScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	result := &scanner.ScanResult{
		ScannerName: n.Name(),
	}

	manager := n.detectManager(ctx)

	var versions []string
	var defaultVersion string

	switch manager {
	case "nvm":
		versions, defaultVersion = n.scanNvm(ctx)
	case "fnm":
		versions, defaultVersion = n.scanFnm(ctx)
	default:
		// Try system node.
		ver, err := n.cmd.Run(ctx, "node", "--version")
		if err != nil {
			// No node installed at all.
			return result, nil
		}
		versions = []string{ver}
		defaultVersion = ver
	}

	if len(versions) == 0 {
		return result, nil
	}

	globalPkgs := n.scanGlobalPackages(ctx)

	result.Node = &domain.NodeSection{
		Manager:        manager,
		Versions:       versions,
		DefaultVersion: defaultVersion,
		GlobalPackages: globalPkgs,
	}
	return result, nil
}

// detectManager checks for nvm (via ~/.nvm/ directory) or fnm (via IsInstalled).
func (n *NodeScanner) detectManager(ctx context.Context) string {
	nvmDir := filepath.Join(n.homeDir, ".nvm")
	if info, err := os.Stat(nvmDir); err == nil && info.IsDir() {
		return "nvm"
	}
	if n.cmd.IsInstalled(ctx, "fnm") {
		return "fnm"
	}
	return ""
}

// scanNvm parses `nvm list` output to extract versions and the default version.
func (n *NodeScanner) scanNvm(ctx context.Context) (versions []string, defaultVersion string) {
	output, err := n.cmd.Run(ctx, "bash", "-l", "-c", "nvm list")
	if err != nil {
		return nil, ""
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Detect default version from "default -> vX.Y.Z" line.
		if strings.HasPrefix(line, "default") {
			parts := strings.Split(line, "->")
			if len(parts) == 2 {
				defaultVersion = strings.TrimSpace(parts[1])
			}
			continue
		}

		// Lines like "->     v20.10.0" or "       v18.19.0"
		cleaned := strings.TrimPrefix(line, "->")
		cleaned = strings.TrimSpace(cleaned)
		if strings.HasPrefix(cleaned, "v") {
			// Take only the version token (first word).
			ver := strings.Fields(cleaned)[0]
			versions = append(versions, ver)
			// The arrow marker also indicates current/default.
			if strings.HasPrefix(line, "->") && defaultVersion == "" {
				defaultVersion = ver
			}
		}
	}
	return versions, defaultVersion
}

// scanFnm parses `fnm list` output to extract versions and the default version.
func (n *NodeScanner) scanFnm(ctx context.Context) (versions []string, defaultVersion string) {
	output, err := n.cmd.Run(ctx, "fnm", "list")
	if err != nil {
		return nil, ""
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// fnm list output: "* v20.11.0 default" or "  v18.19.1"
		isDefault := strings.HasPrefix(line, "*")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		ver := fields[0]
		if strings.HasPrefix(ver, "v") {
			versions = append(versions, ver)
			if isDefault {
				defaultVersion = ver
			}
		}
	}
	return versions, defaultVersion
}

// npmListOutput represents the JSON structure of `npm list -g --depth=0 --json`.
type npmListOutput struct {
	Dependencies map[string]struct {
		Version string `json:"version"`
	} `json:"dependencies"`
}

// scanGlobalPackages parses `npm list -g --depth=0 --json` output.
func (n *NodeScanner) scanGlobalPackages(ctx context.Context) []domain.Package {
	output, err := n.cmd.Run(ctx, "npm", "list", "-g", "--depth=0", "--json")
	if err != nil {
		return nil
	}

	var parsed npmListOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return nil
	}

	// Collect and sort package names for deterministic output.
	names := make([]string, 0, len(parsed.Dependencies))
	for name := range parsed.Dependencies {
		// Skip npm itself — it comes bundled with node.
		if name == "npm" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	var pkgs []domain.Package
	for _, name := range names {
		dep := parsed.Dependencies[name]
		pkgs = append(pkgs, domain.Package{
			Name:    name,
			Version: dep.Version,
		})
	}
	return pkgs
}
