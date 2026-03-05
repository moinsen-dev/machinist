package editors

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJetBrainsScanner_Name(t *testing.T) {
	s := NewJetBrainsScanner("/tmp")
	assert.Equal(t, "jetbrains", s.Name())
}

func TestJetBrainsScanner_Description(t *testing.T) {
	s := NewJetBrainsScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestJetBrainsScanner_Category(t *testing.T) {
	s := NewJetBrainsScanner("/tmp")
	assert.Equal(t, "editors", s.Category())
}

func TestJetBrainsScanner_Scan_NoDirExists(t *testing.T) {
	homeDir := t.TempDir()
	// No JetBrains directory created — scanner must return nil section.
	s := NewJetBrainsScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.JetBrains)
}

func TestJetBrainsScanner_Scan_EmptyJetBrainsDir(t *testing.T) {
	homeDir := t.TempDir()
	jetbrainsDir := filepath.Join(homeDir, "Library", "Application Support", "JetBrains")
	require.NoError(t, os.MkdirAll(jetbrainsDir, 0755))

	s := NewJetBrainsScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.JetBrains)
}

func TestJetBrainsScanner_Scan_KnownIDEs(t *testing.T) {
	homeDir := t.TempDir()
	jetbrainsDir := filepath.Join(homeDir, "Library", "Application Support", "JetBrains")
	require.NoError(t, os.MkdirAll(jetbrainsDir, 0755))

	// Create fake IDE settings directories with versioned suffixes (as JetBrains does).
	ideDirs := []string{
		"IntelliJIdea2023.3",
		"PyCharm2023.3",
		"WebStorm2024.1",
		"GoLand2023.3",
	}
	for _, dir := range ideDirs {
		require.NoError(t, os.MkdirAll(filepath.Join(jetbrainsDir, dir), 0755))
	}

	s := NewJetBrainsScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.JetBrains)
	assert.Len(t, result.JetBrains.IDEs, 4)

	// Collect display names for assertion.
	names := make(map[string]bool)
	for _, ide := range result.JetBrains.IDEs {
		names[ide.Name] = true
		// SettingsExport must point to the actual directory.
		assert.DirExists(t, ide.SettingsExport)
	}
	assert.True(t, names["IntelliJ IDEA"])
	assert.True(t, names["PyCharm Professional"])
	assert.True(t, names["WebStorm"])
	assert.True(t, names["GoLand"])
}

func TestJetBrainsScanner_Scan_UnknownDirsIgnored(t *testing.T) {
	homeDir := t.TempDir()
	jetbrainsDir := filepath.Join(homeDir, "Library", "Application Support", "JetBrains")
	require.NoError(t, os.MkdirAll(jetbrainsDir, 0755))

	// Mix of known and unknown directories.
	require.NoError(t, os.MkdirAll(filepath.Join(jetbrainsDir, "CLion2023.3"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(jetbrainsDir, "SomeRandomDir"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(jetbrainsDir, "Toolbox"), 0755))

	s := NewJetBrainsScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.JetBrains)
	// Only CLion should be detected.
	assert.Len(t, result.JetBrains.IDEs, 1)
	assert.Equal(t, "CLion", result.JetBrains.IDEs[0].Name)
}

func TestJetBrainsScanner_Scan_AllKnownPrefixes(t *testing.T) {
	homeDir := t.TempDir()
	jetbrainsDir := filepath.Join(homeDir, "Library", "Application Support", "JetBrains")
	require.NoError(t, os.MkdirAll(jetbrainsDir, 0755))

	allIDEDirs := []string{
		"IntelliJIdea2024.1",
		"PyCharm2024.1",
		"WebStorm2024.1",
		"GoLand2024.1",
		"AndroidStudio2024.1",
		"PhpStorm2024.1",
		"CLion2024.1",
		"Rider2024.1",
		"DataGrip2024.1",
		"RubyMine2024.1",
	}
	for _, dir := range allIDEDirs {
		require.NoError(t, os.MkdirAll(filepath.Join(jetbrainsDir, dir), 0755))
	}

	s := NewJetBrainsScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.JetBrains)
	assert.Len(t, result.JetBrains.IDEs, 10)
}

func TestJetBrainsScanner_Scan_SettingsExportPath(t *testing.T) {
	homeDir := t.TempDir()
	jetbrainsDir := filepath.Join(homeDir, "Library", "Application Support", "JetBrains")
	ideDir := filepath.Join(jetbrainsDir, "RubyMine2023.3")
	require.NoError(t, os.MkdirAll(ideDir, 0755))

	s := NewJetBrainsScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.JetBrains)
	require.Len(t, result.JetBrains.IDEs, 1)
	assert.Equal(t, "RubyMine", result.JetBrains.IDEs[0].Name)
	assert.Equal(t, ideDir, result.JetBrains.IDEs[0].SettingsExport)
}
