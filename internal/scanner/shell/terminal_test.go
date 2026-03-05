package shell

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalScanner_Name(t *testing.T) {
	s := NewTerminalScanner("/tmp")
	assert.Equal(t, "terminal", s.Name())
}

func TestTerminalScanner_Description(t *testing.T) {
	s := NewTerminalScanner("/tmp")
	assert.NotEmpty(t, s.Description())
}

func TestTerminalScanner_Category(t *testing.T) {
	s := NewTerminalScanner("/tmp")
	assert.Equal(t, "shell", s.Category())
}

func TestTerminalScanner_Scan_NoConfig(t *testing.T) {
	homeDir := t.TempDir()

	s := NewTerminalScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Terminal, "Terminal section should be nil when no config is found")
}

func TestTerminalScanner_Scan_ITerm2(t *testing.T) {
	homeDir := t.TempDir()

	// Create the iTerm2 plist.
	prefsDir := filepath.Join(homeDir, "Library", "Preferences")
	require.NoError(t, os.MkdirAll(prefsDir, 0755))
	plistPath := filepath.Join(prefsDir, "com.googlecode.iterm2.plist")
	require.NoError(t, os.WriteFile(plistPath, []byte("<?xml version=\"1.0\"?>"), 0644))

	s := NewTerminalScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Terminal)
	assert.Equal(t, "iTerm2", result.Terminal.App)
	require.Len(t, result.Terminal.ConfigFiles, 1)
	assert.Equal(t, "Library/Preferences/com.googlecode.iterm2.plist", result.Terminal.ConfigFiles[0].Source)
	assert.NotEmpty(t, result.Terminal.ConfigFiles[0].ContentHash)
}

func TestTerminalScanner_Scan_Warp(t *testing.T) {
	homeDir := t.TempDir()

	// Create ~/.warp directory.
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".warp"), 0755))

	s := NewTerminalScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Terminal)
	assert.Equal(t, "Warp", result.Terminal.App)
	require.Len(t, result.Terminal.ConfigFiles, 1)
	assert.Equal(t, ".warp", result.Terminal.ConfigFiles[0].Source)
}

func TestTerminalScanner_Scan_Alacritty_Toml(t *testing.T) {
	homeDir := t.TempDir()

	alacrittyDir := filepath.Join(homeDir, ".config", "alacritty")
	require.NoError(t, os.MkdirAll(alacrittyDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(alacrittyDir, "alacritty.toml"), []byte("[window]\nopacity = 1.0\n"), 0644))

	s := NewTerminalScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Terminal)
	assert.Equal(t, "Alacritty", result.Terminal.App)
	require.Len(t, result.Terminal.ConfigFiles, 1)
	assert.Equal(t, ".config/alacritty/alacritty.toml", result.Terminal.ConfigFiles[0].Source)
	assert.NotEmpty(t, result.Terminal.ConfigFiles[0].ContentHash)
}

func TestTerminalScanner_Scan_Alacritty_Yml(t *testing.T) {
	homeDir := t.TempDir()

	alacrittyDir := filepath.Join(homeDir, ".config", "alacritty")
	require.NoError(t, os.MkdirAll(alacrittyDir, 0755))
	// Only the .yml file present (no .toml).
	require.NoError(t, os.WriteFile(filepath.Join(alacrittyDir, "alacritty.yml"), []byte("window:\n  opacity: 1.0\n"), 0644))

	s := NewTerminalScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Terminal)
	assert.Equal(t, "Alacritty", result.Terminal.App)
	require.Len(t, result.Terminal.ConfigFiles, 1)
	assert.Equal(t, ".config/alacritty/alacritty.yml", result.Terminal.ConfigFiles[0].Source)
}

func TestTerminalScanner_Scan_WezTerm(t *testing.T) {
	homeDir := t.TempDir()

	weztermDir := filepath.Join(homeDir, ".config", "wezterm")
	require.NoError(t, os.MkdirAll(weztermDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(weztermDir, "wezterm.lua"), []byte("local wezterm = require 'wezterm'\n"), 0644))

	s := NewTerminalScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Terminal)
	assert.Equal(t, "WezTerm", result.Terminal.App)
	require.Len(t, result.Terminal.ConfigFiles, 1)
	assert.Equal(t, ".config/wezterm/wezterm.lua", result.Terminal.ConfigFiles[0].Source)
	assert.NotEmpty(t, result.Terminal.ConfigFiles[0].ContentHash)
}

// TestTerminalScanner_Scan_FirstMatchWins ensures iTerm2 wins when multiple
// terminal configs are present simultaneously.
func TestTerminalScanner_Scan_FirstMatchWins(t *testing.T) {
	homeDir := t.TempDir()

	// iTerm2
	prefsDir := filepath.Join(homeDir, "Library", "Preferences")
	require.NoError(t, os.MkdirAll(prefsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(prefsDir, "com.googlecode.iterm2.plist"), []byte("plist"), 0644))

	// Warp also present
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".warp"), 0755))

	s := NewTerminalScanner(homeDir)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Terminal)
	// iTerm2 is first in the candidate list.
	assert.Equal(t, "iTerm2", result.Terminal.App)
}
