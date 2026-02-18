package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/moinsen-dev/machinist/internal/bundler"
	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/moinsen-dev/machinist/profiles"
)

// MachinistServer wraps the MCP server with the machinist scanner registry.
type MachinistServer struct {
	registry *scanner.Registry
	server   *server.MCPServer
	handlers map[string]server.ToolHandlerFunc
}

// NewMachinistServer creates a new MCP server with all machinist tools registered.
func NewMachinistServer(registry *scanner.Registry) *MachinistServer {
	s := &MachinistServer{
		registry: registry,
		server:   server.NewMCPServer("machinist", "0.1.0", server.WithResourceCapabilities(false, false)),
		handlers: make(map[string]server.ToolHandlerFunc),
	}

	s.registerTools()
	s.registerResources()
	return s
}

// MCPServer returns the underlying MCP server instance.
func (s *MachinistServer) MCPServer() *server.MCPServer {
	return s.server
}

func (s *MachinistServer) registerTools() {
	s.addTool("list_scanners",
		gomcp.NewTool("list_scanners",
			gomcp.WithDescription("List all available scanners"),
		),
		s.handleListScanners,
	)

	s.addTool("scan",
		gomcp.NewTool("scan",
			gomcp.WithDescription("Run a single scanner by name"),
			gomcp.WithString("scanner",
				gomcp.Required(),
				gomcp.Description("Name of the scanner to run"),
			),
		),
		s.handleScan,
	)

	s.addTool("scan_all",
		gomcp.NewTool("scan_all",
			gomcp.WithDescription("Run all scanners and return the full TOML manifest"),
		),
		s.handleScanAll,
	)

	s.addTool("list_profiles",
		gomcp.NewTool("list_profiles",
			gomcp.WithDescription("List all available profile names"),
		),
		s.handleListProfiles,
	)

	s.addTool("get_profile",
		gomcp.NewTool("get_profile",
			gomcp.WithDescription("Get a profile by name as TOML"),
			gomcp.WithString("name",
				gomcp.Required(),
				gomcp.Description("Name of the profile"),
			),
		),
		s.handleGetProfile,
	)

	s.addTool("compose_manifest",
		gomcp.NewTool("compose_manifest",
			gomcp.WithDescription("Compose a TOML manifest from a base profile with optional additional packages"),
			gomcp.WithString("base_profile",
				gomcp.Required(),
				gomcp.Description("Base profile name to start from"),
			),
			gomcp.WithArray("add_packages",
				gomcp.Description("Additional Homebrew formulae to add"),
				gomcp.WithStringItems(),
			),
		),
		s.handleComposeManifest,
	)

	s.addTool("validate_manifest",
		gomcp.NewTool("validate_manifest",
			gomcp.WithDescription("Validate a TOML manifest by attempting to unmarshal it into a Snapshot"),
			gomcp.WithString("manifest",
				gomcp.Required(),
				gomcp.Description("Raw TOML content to validate"),
			),
		),
		s.handleValidateManifest,
	)

	s.addTool("build_dmg",
		gomcp.NewTool("build_dmg",
			gomcp.WithDescription("Build a DMG disk image from a TOML manifest"),
			gomcp.WithString("manifest",
				gomcp.Required(),
				gomcp.Description("Raw TOML manifest content"),
			),
			gomcp.WithString("output_path",
				gomcp.Description("Output path for the DMG file (default: machinist-setup.dmg)"),
			),
			gomcp.WithString("password",
				gomcp.Description("Optional password for AES-256 encryption"),
			),
			gomcp.WithString("volume_name",
				gomcp.Description("Volume name for the DMG (default: machinist)"),
			),
		),
		s.handleBuildDMG,
	)

	s.addTool("diff_manifests",
		gomcp.NewTool("diff_manifests",
			gomcp.WithDescription("Compare two TOML manifests and return a human-readable diff summary"),
			gomcp.WithString("manifest_a",
				gomcp.Required(),
				gomcp.Description("First TOML manifest"),
			),
			gomcp.WithString("manifest_b",
				gomcp.Required(),
				gomcp.Description("Second TOML manifest"),
			),
		),
		s.handleDiffManifests,
	)
}

func (s *MachinistServer) addTool(name string, tool gomcp.Tool, handler server.ToolHandlerFunc) {
	s.handlers[name] = handler
	s.server.AddTool(tool, handler)
}

// handleListScanners returns a JSON array of scanner info.
func (s *MachinistServer) handleListScanners(_ context.Context, _ gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	scanners := s.registry.List()

	type scannerInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Category    string `json:"category"`
	}

	infos := make([]scannerInfo, len(scanners))
	for i, sc := range scanners {
		infos[i] = scannerInfo{
			Name:        sc.Name(),
			Description: sc.Description(),
			Category:    sc.Category(),
		}
	}

	data, err := json.Marshal(infos)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to marshal scanners: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// handleScan runs a single scanner and returns its result as TOML.
func (s *MachinistServer) handleScan(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	name, err := req.RequireString("scanner")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	result, err := s.registry.ScanOne(ctx, name)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("scan failed: %v", err)), nil
	}

	// Build a minimal snapshot with just this result
	snap := domain.NewSnapshot("mcp-scan", runtime.GOOS, runtime.GOARCH, "0.1.0")
	applyResult(snap, result)

	data, err := domain.MarshalManifest(snap)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to marshal manifest: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// handleScanAll runs all scanners and returns the full TOML manifest.
func (s *MachinistServer) handleScanAll(ctx context.Context, _ gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	snap, errs := s.registry.ScanAll(ctx)
	if len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Error()
		}
		// Still return the partial result but note errors
		data, marshalErr := domain.MarshalManifest(snap)
		if marshalErr != nil {
			return gomcp.NewToolResultError(fmt.Sprintf("scan errors: %s; marshal error: %v",
				strings.Join(msgs, "; "), marshalErr)), nil
		}
		return gomcp.NewToolResultText(fmt.Sprintf("# Warnings: %s\n\n%s",
			strings.Join(msgs, "; "), string(data))), nil
	}

	data, err := domain.MarshalManifest(snap)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to marshal manifest: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// handleListProfiles returns a JSON array of profile names.
func (s *MachinistServer) handleListProfiles(_ context.Context, _ gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	profilesDir := profilesDirectory()
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to read profiles directory: %v", err)), nil
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".toml") {
			names = append(names, strings.TrimSuffix(name, ".toml"))
		}
	}
	sort.Strings(names)

	data, err := json.Marshal(names)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to marshal profile names: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// handleGetProfile returns the TOML content of a named profile.
func (s *MachinistServer) handleGetProfile(_ context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	profilePath := filepath.Join(profilesDirectory(), name+".toml")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("profile %q not found: %v", name, err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// handleComposeManifest loads a base profile and appends optional packages.
func (s *MachinistServer) handleComposeManifest(_ context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	baseProfile, err := req.RequireString("base_profile")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	profilePath := filepath.Join(profilesDirectory(), baseProfile+".toml")
	snap, err := domain.ReadManifest(profilePath)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to load profile %q: %v", baseProfile, err)), nil
	}

	addPkgs := req.GetStringSlice("add_packages", nil)
	if len(addPkgs) > 0 {
		if snap.Homebrew == nil {
			snap.Homebrew = &domain.HomebrewSection{}
		}
		for _, pkg := range addPkgs {
			snap.Homebrew.Formulae = append(snap.Homebrew.Formulae, domain.Package{Name: pkg})
		}
	}

	data, err := domain.MarshalManifest(snap)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to marshal manifest: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// handleValidateManifest validates a TOML manifest by attempting to unmarshal it.
func (s *MachinistServer) handleValidateManifest(_ context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	manifest, err := req.RequireString("manifest")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	snap, err := domain.UnmarshalManifest([]byte(manifest))
	if err != nil {
		result := map[string]interface{}{
			"valid": false,
			"error": fmt.Sprintf("parse error: %v", err),
		}
		data, _ := json.Marshal(result)
		return gomcp.NewToolResultText(string(data)), nil
	}

	sections := populatedSections(snap)
	result := map[string]interface{}{
		"valid":    true,
		"sections": sections,
	}
	data, err := json.Marshal(result)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// handleBuildDMG builds a DMG disk image from a TOML manifest.
func (s *MachinistServer) handleBuildDMG(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	manifest, err := req.RequireString("manifest")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	snap, err := domain.UnmarshalManifest([]byte(manifest))
	if err != nil {
		result := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("invalid manifest: %v", err),
		}
		data, _ := json.Marshal(result)
		return gomcp.NewToolResultText(string(data)), nil
	}

	outputPath := req.GetString("output_path", "machinist-setup.dmg")
	volumeName := req.GetString("volume_name", "machinist")
	password := req.GetString("password", "")

	cmd := &util.RealCommandRunner{}
	opts := bundler.BundleOptions{
		Password:   password,
		VolumeName: volumeName,
	}

	if err := bundler.Bundle(ctx, cmd, snap, outputPath, opts); err != nil {
		result := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("bundle failed: %v", err),
		}
		data, _ := json.Marshal(result)
		return gomcp.NewToolResultText(string(data)), nil
	}

	result := map[string]interface{}{
		"success": true,
		"path":    outputPath,
	}
	data, err := json.Marshal(result)
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return gomcp.NewToolResultText(string(data)), nil
}

// handleDiffManifests compares two TOML manifests and returns a human-readable diff.
func (s *MachinistServer) handleDiffManifests(_ context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
	manifestA, err := req.RequireString("manifest_a")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}
	manifestB, err := req.RequireString("manifest_b")
	if err != nil {
		return gomcp.NewToolResultError(err.Error()), nil
	}

	snapA, err := domain.UnmarshalManifest([]byte(manifestA))
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to parse manifest_a: %v", err)), nil
	}
	snapB, err := domain.UnmarshalManifest([]byte(manifestB))
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("failed to parse manifest_b: %v", err)), nil
	}

	sectionsA := populatedSections(snapA)
	sectionsB := populatedSections(snapB)

	setA := make(map[string]bool, len(sectionsA))
	for _, s := range sectionsA {
		setA[s] = true
	}
	setB := make(map[string]bool, len(sectionsB))
	for _, s := range sectionsB {
		setB[s] = true
	}

	var onlyA, onlyB, both []string
	for _, s := range sectionsA {
		if setB[s] {
			both = append(both, s)
		} else {
			onlyA = append(onlyA, s)
		}
	}
	for _, s := range sectionsB {
		if !setA[s] {
			onlyB = append(onlyB, s)
		}
	}

	sort.Strings(onlyA)
	sort.Strings(onlyB)
	sort.Strings(both)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Only in A: %v\n", onlyA))
	sb.WriteString(fmt.Sprintf("Only in B: %v\n", onlyB))
	sb.WriteString(fmt.Sprintf("In both: %v\n", both))

	// Homebrew package-level diff
	if snapA.Homebrew != nil && snapB.Homebrew != nil {
		pkgsA := make(map[string]bool)
		for _, p := range snapA.Homebrew.Formulae {
			pkgsA[p.Name] = true
		}
		for _, p := range snapA.Homebrew.Casks {
			pkgsA[p.Name] = true
		}
		pkgsB := make(map[string]bool)
		for _, p := range snapB.Homebrew.Formulae {
			pkgsB[p.Name] = true
		}
		for _, p := range snapB.Homebrew.Casks {
			pkgsB[p.Name] = true
		}

		var hOnlyA, hOnlyB []string
		for name := range pkgsA {
			if !pkgsB[name] {
				hOnlyA = append(hOnlyA, name)
			}
		}
		for name := range pkgsB {
			if !pkgsA[name] {
				hOnlyB = append(hOnlyB, name)
			}
		}
		sort.Strings(hOnlyA)
		sort.Strings(hOnlyB)

		if len(hOnlyA) > 0 || len(hOnlyB) > 0 {
			sb.WriteString("\nHomebrew diff:\n")
			if len(hOnlyA) > 0 {
				sb.WriteString(fmt.Sprintf("  Only in A: %v\n", hOnlyA))
			}
			if len(hOnlyB) > 0 {
				sb.WriteString(fmt.Sprintf("  Only in B: %v\n", hOnlyB))
			}
		}
	}

	return gomcp.NewToolResultText(sb.String()), nil
}

// populatedSections returns the TOML tag names of non-nil pointer sections in a Snapshot.
func populatedSections(snap *domain.Snapshot) []string {
	var sections []string
	v := reflect.ValueOf(*snap)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr && !field.IsNil() {
			tag := t.Field(i).Tag.Get("toml")
			if tag == "" {
				continue
			}
			// Strip ",omitempty" or similar suffixes
			if idx := strings.Index(tag, ","); idx != -1 {
				tag = tag[:idx]
			}
			sections = append(sections, tag)
		}
	}
	return sections
}

// registerResources registers MCP resources for system snapshot and profiles.
func (s *MachinistServer) registerResources() {
	// Static resource: system snapshot
	s.server.AddResource(
		gomcp.NewResource(
			"machinist://system/snapshot",
			"System Snapshot",
			gomcp.WithResourceDescription("Returns the current system snapshot as TOML by running all scanners"),
			gomcp.WithMIMEType("application/toml"),
		),
		s.handleSnapshotResource,
	)

	// Template resource: profiles by name
	s.server.AddResourceTemplate(
		gomcp.NewResourceTemplate(
			"machinist://profiles/{name}",
			"Profile",
			gomcp.WithTemplateDescription("Returns a specific profile TOML by name"),
			gomcp.WithTemplateMIMEType("application/toml"),
		),
		s.handleProfileResource,
	)
}

// handleSnapshotResource returns the current system snapshot as TOML.
func (s *MachinistServer) handleSnapshotResource(ctx context.Context, req gomcp.ReadResourceRequest) ([]gomcp.ResourceContents, error) {
	snap, errs := s.registry.ScanAll(ctx)
	if len(errs) > 0 && snap == nil {
		return nil, fmt.Errorf("scan failed: %v", errs[0])
	}

	data, err := domain.MarshalManifest(snap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}

	return []gomcp.ResourceContents{
		gomcp.TextResourceContents{
			URI:      "machinist://system/snapshot",
			MIMEType: "application/toml",
			Text:     string(data),
		},
	}, nil
}

// handleProfileResource returns a specific profile TOML by name.
func (s *MachinistServer) handleProfileResource(_ context.Context, req gomcp.ReadResourceRequest) ([]gomcp.ResourceContents, error) {
	uri := req.Params.URI
	// Extract name from URI: machinist://profiles/{name}
	name := strings.TrimPrefix(uri, "machinist://profiles/")
	if name == "" || name == uri {
		return nil, fmt.Errorf("invalid profile URI: %s", uri)
	}

	snap, err := profiles.Get(name)
	if err != nil {
		return nil, fmt.Errorf("profile %q not found: %w", name, err)
	}

	data, err := domain.MarshalManifest(snap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile: %w", err)
	}

	return []gomcp.ResourceContents{
		gomcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/toml",
			Text:     string(data),
		},
	}, nil
}

// profilesDirectory returns the path to the profiles directory.
func profilesDirectory() string {
	// Walk up from the executable or use a known relative path.
	// For development, find relative to the source tree.
	// Try finding the profiles dir relative to the working directory.
	if dir, err := findProjectRoot(); err == nil {
		return filepath.Join(dir, "profiles")
	}
	return "profiles"
}

// findProjectRoot walks up from the working directory looking for go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("project root not found")
		}
		dir = parent
	}
}

// applyResult maps a ScanResult's populated fields onto the Snapshot.
func applyResult(snap *domain.Snapshot, result *scanner.ScanResult) {
	if result.Homebrew != nil {
		snap.Homebrew = result.Homebrew
	}
	if result.Shell != nil {
		snap.Shell = result.Shell
	}
	if result.Node != nil {
		snap.Node = result.Node
	}
	if result.Python != nil {
		snap.Python = result.Python
	}
	if result.Rust != nil {
		snap.Rust = result.Rust
	}
	if result.Git != nil {
		snap.Git = result.Git
	}
	if result.GitRepos != nil {
		snap.GitRepos = result.GitRepos
	}
	if result.VSCode != nil {
		snap.VSCode = result.VSCode
	}
	if result.Cursor != nil {
		snap.Cursor = result.Cursor
	}
	if result.Docker != nil {
		snap.Docker = result.Docker
	}
	if result.MacOSDefaults != nil {
		snap.MacOSDefaults = result.MacOSDefaults
	}
	if result.Folders != nil {
		snap.Folders = result.Folders
	}
	if result.Fonts != nil {
		snap.Fonts = result.Fonts
	}
	if result.Crontab != nil {
		snap.Crontab = result.Crontab
	}
	if result.LaunchAgents != nil {
		snap.LaunchAgents = result.LaunchAgents
	}
	if result.Apps != nil {
		snap.Apps = result.Apps
	}
}
