package system

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFoldersScanner_Name(t *testing.T) {
	s := NewFoldersScanner("/tmp")
	assert.Equal(t, "folders", s.Name())
}

func TestFoldersScanner_Category(t *testing.T) {
	s := NewFoldersScanner("/tmp")
	assert.Equal(t, "system", s.Category())
}

func TestFoldersScanner_Scan_TopLevelDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirs including a hidden one.
	for _, name := range []string{"Code", "Projects", "Workspace", ".hidden"} {
		require.NoError(t, os.Mkdir(filepath.Join(tmpDir, name), 0o755))
	}

	s := NewFoldersScanner(tmpDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Folders)

	dirs := result.Folders.Structure
	assert.Contains(t, dirs, "Code")
	assert.Contains(t, dirs, "Projects")
	assert.Contains(t, dirs, "Workspace")
	assert.NotContains(t, dirs, ".hidden")
}

func TestFoldersScanner_Scan_EmptyHome(t *testing.T) {
	tmpDir := t.TempDir()

	s := NewFoldersScanner(tmpDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Folders)
	assert.Empty(t, result.Folders.Structure)
}

func TestFoldersScanner_Scan_SkipsSystemDirs(t *testing.T) {
	tmpDir := t.TempDir()

	for _, name := range []string{"Library", "Applications", "Code", ".Trash"} {
		require.NoError(t, os.Mkdir(filepath.Join(tmpDir, name), 0o755))
	}

	s := NewFoldersScanner(tmpDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Folders)

	dirs := result.Folders.Structure
	assert.Equal(t, []string{"Code"}, dirs)
}
