package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBetterTouchToolScanner_Name(t *testing.T) {
	s := NewBetterTouchToolScanner("/tmp")
	assert.Equal(t, "bettertouchtool", s.Name())
}

func TestBetterTouchToolScanner_Description(t *testing.T) {
	s := NewBetterTouchToolScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestBetterTouchToolScanner_Category(t *testing.T) {
	s := NewBetterTouchToolScanner("/tmp")
	assert.Equal(t, "tools", s.Category())
}

func TestBetterTouchToolScanner_Scan_NoDirFound(t *testing.T) {
	homeDir := t.TempDir()
	s := NewBetterTouchToolScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.BetterTouchTool)
}

func TestBetterTouchToolScanner_Scan_DirExists(t *testing.T) {
	homeDir := t.TempDir()
	bttDir := filepath.Join(homeDir, "Library", "Application Support", "BetterTouchTool")
	require.NoError(t, os.MkdirAll(bttDir, 0o755))

	s := NewBetterTouchToolScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.BetterTouchTool)
	assert.Equal(t, bttDir, result.BetterTouchTool.ConfigFile)
}
