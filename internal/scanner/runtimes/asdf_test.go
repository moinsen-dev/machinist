package runtimes

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsdfScanner_Name(t *testing.T) {
	s := NewAsdfScanner("/home/user", &util.MockCommandRunner{})
	assert.Equal(t, "asdf", s.Name())
}

func TestAsdfScanner_Description(t *testing.T) {
	s := NewAsdfScanner("/home/user", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestAsdfScanner_Category(t *testing.T) {
	s := NewAsdfScanner("/home/user", &util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

// TestAsdfScanner_Scan_HappyPath_Asdf tests a full asdf scan with two plugins
// and multiple versions per plugin, including leading spaces and '*' markers.
func TestAsdfScanner_Scan_HappyPath_Asdf(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// IsInstalled("asdf") → true
			"asdf": {Output: "", Err: nil},
			// Plugin list
			"asdf plugin list": {Output: "nodejs\npython"},
			// Versions for nodejs: one active (marked with *), one inactive
			"asdf list nodejs": {Output: "  * 20.11.0\n  18.19.1"},
			// Versions for python: two versions, first is active
			"asdf list python": {Output: "  * 3.12.1\n  3.11.5\n  3.10.13"},
		},
	}

	s := NewAsdfScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Asdf)

	asdf := result.Asdf
	require.Len(t, asdf.Plugins, 2)

	// nodejs plugin
	nodePlugin := asdf.Plugins[0]
	assert.Equal(t, "nodejs", nodePlugin.Name)
	assert.Equal(t, []string{"20.11.0", "18.19.1"}, nodePlugin.Versions)

	// python plugin
	pyPlugin := asdf.Plugins[1]
	assert.Equal(t, "python", pyPlugin.Name)
	assert.Equal(t, []string{"3.12.1", "3.11.5", "3.10.13"}, pyPlugin.Versions)

	// No .tool-versions file in the temp dir, so ToolVersionsFile should be empty.
	assert.Empty(t, asdf.ToolVersionsFile)
}

// TestAsdfScanner_Scan_WithToolVersionsFile tests that the .tool-versions file path
// is recorded when it exists in the home directory.
func TestAsdfScanner_Scan_WithToolVersionsFile(t *testing.T) {
	homeDir := t.TempDir()

	// Create a .tool-versions file in the temp home dir.
	toolVersionsPath := filepath.Join(homeDir, ".tool-versions")
	require.NoError(t, os.WriteFile(toolVersionsPath, []byte("nodejs 20.11.0\npython 3.12.1\n"), 0o644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"asdf": {Output: "", Err: nil},
			"asdf plugin list": {Output: "nodejs"},
			"asdf list nodejs": {Output: "  20.11.0"},
		},
	}

	s := NewAsdfScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Asdf)

	assert.Equal(t, toolVersionsPath, result.Asdf.ToolVersionsFile)
}

// TestAsdfScanner_Scan_NeitherInstalled tests that when neither asdf nor mise
// is found, the Asdf section is nil.
func TestAsdfScanner_Scan_NeitherInstalled(t *testing.T) {
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// No "asdf" or "mise" keys → IsInstalled returns false for both.
		},
	}

	s := NewAsdfScanner("/home/user", mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Asdf)
}

// TestAsdfScanner_Scan_HappyPath_Mise tests a full mise scan with two plugins
// and multiple versions parsed from `mise list` output.
func TestAsdfScanner_Scan_HappyPath_Mise(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// IsInstalled("asdf") → false (no "asdf" key)
			// IsInstalled("mise") → true
			"mise": {Output: "", Err: nil},
			// Plugin list
			"mise plugins list": {Output: "node\npython"},
			// All versions in one call (format: tool  version  source)
			"mise list": {
				Output: "node  20.11.0  ~/.tool-versions\nnode  18.19.1  ~/.tool-versions\npython  3.12.1  ~/.tool-versions",
			},
		},
	}

	s := NewAsdfScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Asdf)

	asdf := result.Asdf
	require.Len(t, asdf.Plugins, 2)

	nodePlugin := asdf.Plugins[0]
	assert.Equal(t, "node", nodePlugin.Name)
	assert.Equal(t, []string{"20.11.0", "18.19.1"}, nodePlugin.Versions)

	pyPlugin := asdf.Plugins[1]
	assert.Equal(t, "python", pyPlugin.Name)
	assert.Equal(t, []string{"3.12.1"}, pyPlugin.Versions)
}

// TestAsdfScanner_Scan_AsdfVersionCleaning tests that '*' markers and leading
// spaces are correctly stripped from asdf list output.
func TestAsdfScanner_Scan_AsdfVersionCleaning(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"asdf": {Output: "", Err: nil},
			"asdf plugin list": {Output: "ruby"},
			// Mix of markers: leading spaces, * with space, plain version.
			"asdf list ruby": {Output: "  * 3.3.0\n3.2.2\n  3.1.4"},
		},
	}

	s := NewAsdfScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Asdf)

	rubyPlugin := result.Asdf.Plugins[0]
	assert.Equal(t, "ruby", rubyPlugin.Name)
	assert.Equal(t, []string{"3.3.0", "3.2.2", "3.1.4"}, rubyPlugin.Versions)
}

// TestAsdfScanner_Scan_EmptyPluginList tests that an empty plugin list
// results in an Asdf section with no plugins (but not nil).
func TestAsdfScanner_Scan_EmptyPluginList(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"asdf":             {Output: "", Err: nil},
			"asdf plugin list": {Output: ""},
		},
	}

	s := NewAsdfScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Asdf)
	assert.Empty(t, result.Asdf.Plugins)
}

// TestCleanAsdfVersion is a unit test for the version-cleaning helper.
func TestCleanAsdfVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  * 20.11.0", "20.11.0"},
		{"  20.11.0", "20.11.0"},
		{"20.11.0", "20.11.0"},
		{"* 20.11.0", "20.11.0"},
		{"  *  3.12.1  ", "3.12.1"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanAsdfVersion(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestParseMiseList is a unit test for the mise list output parser.
func TestParseMiseList(t *testing.T) {
	lines := []string{
		"node  20.11.0  ~/.tool-versions",
		"node  18.19.1  ~/.tool-versions",
		"python  3.12.1  system",
		"",
		"  ",
	}

	got := parseMiseList(lines)

	assert.Equal(t, []string{"20.11.0", "18.19.1"}, got["node"])
	assert.Equal(t, []string{"3.12.1"}, got["python"])

	// The domain type should list all plugins, so we verify map construction.
	assert.Len(t, got, 2)
}

// assertAsdfPlugin is a helper to find a plugin by name and assert its versions.
func assertAsdfPlugin(t *testing.T, plugins []domain.AsdfPlugin, name string, versions []string) {
	t.Helper()
	for _, p := range plugins {
		if p.Name == name {
			assert.Equal(t, versions, p.Versions, "version mismatch for plugin %s", name)
			return
		}
	}
	t.Errorf("plugin %s not found in %v", name, plugins)
}
