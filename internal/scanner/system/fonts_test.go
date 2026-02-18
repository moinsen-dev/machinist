package system

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFontsScanner_Name(t *testing.T) {
	s := NewFontsScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "fonts", s.Name())
}

func TestFontsScanner_Category(t *testing.T) {
	s := NewFontsScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "system", s.Category())
}

func TestFontsScanner_Scan_UserFonts(t *testing.T) {
	tmpHome := t.TempDir()
	fontsDir := filepath.Join(tmpHome, "Library", "Fonts")
	require.NoError(t, os.MkdirAll(fontsDir, 0o755))

	// Create font files and a non-font file
	for _, name := range []string{"MyFont.ttf", "CustomFont.otf", "ReadMe.txt"} {
		require.NoError(t, os.WriteFile(filepath.Join(fontsDir, name), []byte("data"), 0o644))
	}

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"brew":                  {Output: "", Err: nil}, // brew is installed
			"brew list --cask": {Output: "", Err: nil},      // no casks
		},
	}

	s := NewFontsScanner(tmpHome, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Fonts)

	assert.Len(t, result.Fonts.CustomFonts, 2)

	names := make([]string, len(result.Fonts.CustomFonts))
	for i, f := range result.Fonts.CustomFonts {
		names[i] = f.Name
	}
	assert.Contains(t, names, "MyFont")
	assert.Contains(t, names, "CustomFont")
}

func TestFontsScanner_Scan_HomebrewFonts(t *testing.T) {
	tmpHome := t.TempDir()
	fontsDir := filepath.Join(tmpHome, "Library", "Fonts")
	require.NoError(t, os.MkdirAll(fontsDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"brew":             {Output: "", Err: nil},
			"brew list --cask": {Output: "font-fira-code\nfont-jetbrains-mono\nfirefox\niterm2", Err: nil},
		},
	}

	s := NewFontsScanner(tmpHome, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Fonts)

	assert.Equal(t, []string{"font-fira-code", "font-jetbrains-mono"}, result.Fonts.HomebrewFonts)
}

func TestFontsScanner_Scan_Combined(t *testing.T) {
	tmpHome := t.TempDir()
	fontsDir := filepath.Join(tmpHome, "Library", "Fonts")
	require.NoError(t, os.MkdirAll(fontsDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(fontsDir, "Roboto.ttf"), []byte("data"), 0o644))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"brew":             {Output: "", Err: nil},
			"brew list --cask": {Output: "font-fira-code\nfirefox", Err: nil},
		},
	}

	s := NewFontsScanner(tmpHome, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Fonts)

	assert.Len(t, result.Fonts.CustomFonts, 1)
	assert.Equal(t, "Roboto", result.Fonts.CustomFonts[0].Name)
	assert.Equal(t, []string{"font-fira-code"}, result.Fonts.HomebrewFonts)
}

func TestFontsScanner_Scan_NoFonts(t *testing.T) {
	tmpHome := t.TempDir()
	fontsDir := filepath.Join(tmpHome, "Library", "Fonts")
	require.NoError(t, os.MkdirAll(fontsDir, 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"brew":             {Output: "", Err: nil},
			"brew list --cask": {Output: "", Err: nil},
		},
	}

	s := NewFontsScanner(tmpHome, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Fonts)

	assert.Empty(t, result.Fonts.CustomFonts)
	assert.Empty(t, result.Fonts.HomebrewFonts)
}

func TestFontsScanner_Scan_NoBrewInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	fontsDir := filepath.Join(tmpHome, "Library", "Fonts")
	require.NoError(t, os.MkdirAll(fontsDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(fontsDir, "Arial.ttc"), []byte("data"), 0o644))

	// No "brew" key means IsInstalled returns false
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	s := NewFontsScanner(tmpHome, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Fonts)

	assert.Len(t, result.Fonts.CustomFonts, 1)
	assert.Equal(t, "Arial", result.Fonts.CustomFonts[0].Name)
	assert.Empty(t, result.Fonts.HomebrewFonts)
}
