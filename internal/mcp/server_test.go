package mcp

import (
	"context"
	"encoding/json"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
)

// mockScanner implements scanner.Scanner for testing.
type mockScanner struct {
	name        string
	description string
	category    string
	result      *scanner.ScanResult
	err         error
}

func (m *mockScanner) Name() string        { return m.name }
func (m *mockScanner) Description() string  { return m.description }
func (m *mockScanner) Category() string     { return m.category }
func (m *mockScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// callTool builds a CallToolRequest and invokes the handler registered on the MCPServer.
func callTool(s *MachinistServer, name string, args map[string]interface{}) (*gomcp.CallToolResult, error) {
	req := gomcp.CallToolRequest{
		Params: gomcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}

	handler, ok := s.handlers[name]
	if !ok {
		return nil, nil
	}
	return handler(context.Background(), req)
}

func newTestRegistry(scanners ...scanner.Scanner) *scanner.Registry {
	reg := scanner.NewRegistry()
	for _, s := range scanners {
		_ = reg.Register(s)
	}
	return reg
}

func TestListScanners(t *testing.T) {
	reg := newTestRegistry(
		&mockScanner{name: "homebrew", description: "Scan Homebrew packages", category: "packages"},
		&mockScanner{name: "shell", description: "Scan shell config", category: "shell"},
	)
	srv := NewMachinistServer(reg)

	result, err := callTool(srv, "list_scanners", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	text := getTextContent(t, result)

	var scanners []map[string]string
	err = json.Unmarshal([]byte(text), &scanners)
	require.NoError(t, err)
	assert.Len(t, scanners, 2)

	names := make(map[string]bool)
	for _, s := range scanners {
		names[s["name"]] = true
	}
	assert.True(t, names["homebrew"])
	assert.True(t, names["shell"])
}

func TestScan(t *testing.T) {
	reg := newTestRegistry(
		&mockScanner{
			name:        "homebrew",
			description: "Scan Homebrew packages",
			category:    "packages",
			result: &scanner.ScanResult{
				ScannerName: "homebrew",
				Homebrew: &domain.HomebrewSection{
					Taps:     []string{"homebrew/core"},
					Formulae: []domain.Package{{Name: "git"}},
				},
			},
		},
	)
	srv := NewMachinistServer(reg)

	result, err := callTool(srv, "scan", map[string]interface{}{"scanner": "homebrew"})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	text := getTextContent(t, result)
	assert.Contains(t, text, "[homebrew]")
	assert.Contains(t, text, "git")
}

func TestScan_UnknownScanner(t *testing.T) {
	reg := newTestRegistry()
	srv := NewMachinistServer(reg)

	result, err := callTool(srv, "scan", map[string]interface{}{"scanner": "unknown"})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)

	text := getTextContent(t, result)
	assert.Contains(t, text, "unknown")
}

func TestScanAll(t *testing.T) {
	reg := newTestRegistry(
		&mockScanner{
			name:        "homebrew",
			description: "Scan Homebrew",
			category:    "packages",
			result: &scanner.ScanResult{
				ScannerName: "homebrew",
				Homebrew: &domain.HomebrewSection{
					Taps: []string{"homebrew/core"},
				},
			},
		},
		&mockScanner{
			name:        "shell",
			description: "Scan shell",
			category:    "shell",
			result: &scanner.ScanResult{
				ScannerName: "shell",
				Shell: &domain.ShellSection{
					DefaultShell: "/bin/zsh",
				},
			},
		},
	)
	srv := NewMachinistServer(reg)

	result, err := callTool(srv, "scan_all", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	text := getTextContent(t, result)
	assert.Contains(t, text, "[homebrew]")
	assert.Contains(t, text, "[shell]")
}

func TestListProfiles(t *testing.T) {
	srv := NewMachinistServer(scanner.NewRegistry())

	result, err := callTool(srv, "list_profiles", nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	text := getTextContent(t, result)

	var profiles []string
	err = json.Unmarshal([]byte(text), &profiles)
	require.NoError(t, err)
	assert.Greater(t, len(profiles), 0, "should find at least one profile")
	assert.Contains(t, profiles, "minimal")
}

func TestGetProfile(t *testing.T) {
	srv := NewMachinistServer(scanner.NewRegistry())

	result, err := callTool(srv, "get_profile", map[string]interface{}{"name": "minimal"})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	text := getTextContent(t, result)
	assert.Contains(t, text, "[homebrew]")
	assert.Contains(t, text, "git")
}

func TestGetProfile_NotFound(t *testing.T) {
	srv := NewMachinistServer(scanner.NewRegistry())

	result, err := callTool(srv, "get_profile", map[string]interface{}{"name": "nonexistent-profile-xyz"})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)

	text := getTextContent(t, result)
	assert.Contains(t, text, "nonexistent-profile-xyz")
}

func TestComposeManifest(t *testing.T) {
	srv := NewMachinistServer(scanner.NewRegistry())

	result, err := callTool(srv, "compose_manifest", map[string]interface{}{
		"base_profile": "minimal",
		"add_packages": []interface{}{"wget", "curl"},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	text := getTextContent(t, result)
	assert.Contains(t, text, "[homebrew]")
	assert.Contains(t, text, "wget")
	assert.Contains(t, text, "curl")
}

func TestComposeManifest_NoAddPackages(t *testing.T) {
	srv := NewMachinistServer(scanner.NewRegistry())

	result, err := callTool(srv, "compose_manifest", map[string]interface{}{
		"base_profile": "minimal",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	text := getTextContent(t, result)
	assert.Contains(t, text, "[homebrew]")
}

func TestMCPServer(t *testing.T) {
	reg := scanner.NewRegistry()
	srv := NewMachinistServer(reg)
	assert.NotNil(t, srv.MCPServer())
}

// getTextContent extracts the text from the first TextContent in a CallToolResult.
func getTextContent(t *testing.T, result *gomcp.CallToolResult) string {
	t.Helper()
	require.NotEmpty(t, result.Content, "expected at least one content item")
	tc, ok := gomcp.AsTextContent(result.Content[0])
	require.True(t, ok, "expected TextContent, got %T", result.Content[0])
	return tc.Text
}
