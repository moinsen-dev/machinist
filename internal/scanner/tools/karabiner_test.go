package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKarabinerScanner_Name(t *testing.T) {
	s := NewKarabinerScanner("/tmp")
	assert.Equal(t, "karabiner", s.Name())
}

func TestKarabinerScanner_Description(t *testing.T) {
	s := NewKarabinerScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestKarabinerScanner_Category(t *testing.T) {
	s := NewKarabinerScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestKarabinerScanner_Scan_NoDirFound(t *testing.T) {
	homeDir := t.TempDir()
	s := NewKarabinerScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Karabiner)
}

func TestKarabinerScanner_Scan_DirExistsButNoJSON(t *testing.T) {
	homeDir := t.TempDir()
	karabinerDir := filepath.Join(homeDir, ".config", "karabiner")
	require.NoError(t, os.MkdirAll(karabinerDir, 0o755))

	s := NewKarabinerScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Karabiner)
}

func TestKarabinerScanner_Scan_DirAndJSONExist(t *testing.T) {
	homeDir := t.TempDir()
	karabinerDir := filepath.Join(homeDir, ".config", "karabiner")
	require.NoError(t, os.MkdirAll(karabinerDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(karabinerDir, "karabiner.json"), []byte("{}"), 0o644))

	s := NewKarabinerScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Karabiner)
	assert.Equal(t, karabinerDir, result.Karabiner.ConfigDir)
}
