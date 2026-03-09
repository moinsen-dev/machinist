package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlfredScanner_Name(t *testing.T) {
	s := NewAlfredScanner("/tmp")
	assert.Equal(t, "alfred", s.Name())
}

func TestAlfredScanner_Description(t *testing.T) {
	s := NewAlfredScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestAlfredScanner_Category(t *testing.T) {
	s := NewAlfredScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestAlfredScanner_Scan_NoDirFound(t *testing.T) {
	homeDir := t.TempDir()
	s := NewAlfredScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Alfred)
}

func TestAlfredScanner_Scan_PrimaryDir(t *testing.T) {
	homeDir := t.TempDir()
	alfredDir := filepath.Join(homeDir, "Library", "Application Support", "Alfred")
	require.NoError(t, os.MkdirAll(alfredDir, 0o755))

	s := NewAlfredScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Alfred)
	assert.Equal(t, filepath.Join("Library", "Application Support", "Alfred"), result.Alfred.ConfigDir)
}

func TestAlfredScanner_Scan_FallbackPlist(t *testing.T) {
	homeDir := t.TempDir()
	prefsDir := filepath.Join(homeDir, "Library", "Preferences")
	require.NoError(t, os.MkdirAll(prefsDir, 0o755))

	plistPath := filepath.Join(prefsDir, "com.runningwithcrayons.Alfred-Preferences-3.plist")
	require.NoError(t, os.WriteFile(plistPath, []byte("plist-data"), 0o644))

	s := NewAlfredScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Alfred)
	assert.Equal(t, filepath.Join("Library", "Preferences", "com.runningwithcrayons.Alfred-Preferences-3.plist"), result.Alfred.ConfigDir)
}

func TestAlfredScanner_Scan_PrimaryTakesPrecedence(t *testing.T) {
	homeDir := t.TempDir()

	// Create both primary and fallback
	alfredDir := filepath.Join(homeDir, "Library", "Application Support", "Alfred")
	require.NoError(t, os.MkdirAll(alfredDir, 0o755))

	prefsDir := filepath.Join(homeDir, "Library", "Preferences")
	require.NoError(t, os.MkdirAll(prefsDir, 0o755))
	plistPath := filepath.Join(prefsDir, "com.runningwithcrayons.Alfred-Preferences-3.plist")
	require.NoError(t, os.WriteFile(plistPath, []byte("plist-data"), 0o644))

	s := NewAlfredScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Alfred)
	// Primary should take precedence over fallback
	assert.Equal(t, filepath.Join("Library", "Application Support", "Alfred"), result.Alfred.ConfigDir)
}
