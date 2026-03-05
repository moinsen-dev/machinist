package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRaycastScanner_Name(t *testing.T) {
	s := NewRaycastScanner("/tmp")
	assert.Equal(t, "raycast", s.Name())
}

func TestRaycastScanner_Description(t *testing.T) {
	s := NewRaycastScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestRaycastScanner_Category(t *testing.T) {
	s := NewRaycastScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestRaycastScanner_Scan_NoDirFound(t *testing.T) {
	homeDir := t.TempDir()
	s := NewRaycastScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Raycast)
}

func TestRaycastScanner_Scan_DirExists(t *testing.T) {
	homeDir := t.TempDir()
	raycastDir := filepath.Join(homeDir, "Library", "Application Support", "com.raycast.macos")
	require.NoError(t, os.MkdirAll(raycastDir, 0o755))

	s := NewRaycastScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Raycast)
	assert.Equal(t, raycastDir, result.Raycast.ExportFile)
}
