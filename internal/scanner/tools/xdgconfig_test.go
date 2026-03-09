package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXDGConfigScanner_Name(t *testing.T) {
	s := NewXDGConfigScanner("/tmp")
	assert.Equal(t, "xdg-config", s.Name())
}

func TestXDGConfigScanner_Description(t *testing.T) {
	s := NewXDGConfigScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestXDGConfigScanner_Category(t *testing.T) {
	s := NewXDGConfigScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestXDGConfigScanner_Scan_NoConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	s := NewXDGConfigScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.XDGConfig)
}

func TestXDGConfigScanner_Scan_EmptyConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".config"), 0o755))

	s := NewXDGConfigScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	// Empty .config dir with no known tools → nil section (nothing to restore)
	assert.Nil(t, result.XDGConfig)
}

func TestXDGConfigScanner_Scan_KnownToolsDetected(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config")

	// Create some known tool directories
	for _, tool := range []string{"bat", "lazygit", "starship"} {
		require.NoError(t, os.MkdirAll(filepath.Join(configDir, tool), 0o755))
	}
	// Create an unknown directory (should not be detected)
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "some-random-tool"), 0o755))
	// Create a file (should be skipped)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "starship.toml"), []byte(""), 0o644))

	s := NewXDGConfigScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.XDGConfig)
	assert.Len(t, result.XDGConfig.AutoDetected, 3)
	assert.Contains(t, result.XDGConfig.AutoDetected, "bat")
	assert.Contains(t, result.XDGConfig.AutoDetected, "lazygit")
	assert.Contains(t, result.XDGConfig.AutoDetected, "starship")
}

func TestXDGConfigScanner_Scan_AllKnownTools(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config")

	allTools := []string{"bat", "lazygit", "htop", "btop", "aerospace", "gh", "starship", "fish", "kitty", "micro"}
	for _, tool := range allTools {
		require.NoError(t, os.MkdirAll(filepath.Join(configDir, tool), 0o755))
	}

	s := NewXDGConfigScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.XDGConfig)
	assert.Len(t, result.XDGConfig.AutoDetected, 10)
}
