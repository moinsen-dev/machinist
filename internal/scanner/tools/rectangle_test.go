package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRectangleScanner_Name(t *testing.T) {
	s := NewRectangleScanner("/tmp")
	assert.Equal(t, "rectangle", s.Name())
}

func TestRectangleScanner_Description(t *testing.T) {
	s := NewRectangleScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestRectangleScanner_Category(t *testing.T) {
	s := NewRectangleScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestRectangleScanner_Scan_NoFileFound(t *testing.T) {
	homeDir := t.TempDir()
	s := NewRectangleScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Rectangle)
}

func TestRectangleScanner_Scan_FileExists(t *testing.T) {
	homeDir := t.TempDir()
	prefsDir := filepath.Join(homeDir, "Library", "Preferences")
	require.NoError(t, os.MkdirAll(prefsDir, 0o755))

	configFile := filepath.Join(prefsDir, "com.knewton.Rectangle.plist")
	require.NoError(t, os.WriteFile(configFile, []byte("plist-data"), 0o644))

	s := NewRectangleScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Rectangle)
	assert.Equal(t, configFile, result.Rectangle.ConfigFile)
}
