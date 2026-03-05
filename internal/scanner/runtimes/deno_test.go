package runtimes

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDenoScanner_Name(t *testing.T) {
	s := NewDenoScanner("/home/user", &util.MockCommandRunner{})
	assert.Equal(t, "deno", s.Name())
}

func TestDenoScanner_Description(t *testing.T) {
	s := NewDenoScanner("/home/user", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestDenoScanner_Category(t *testing.T) {
	s := NewDenoScanner("/home/user", &util.MockCommandRunner{})
	assert.Equal(t, "runtimes", s.Category())
}

func TestDenoScanner_Scan_HappyPath(t *testing.T) {
	homeDir := t.TempDir()

	// Create fake ~/.deno/bin/ with some installed scripts.
	binDir := filepath.Join(homeDir, ".deno", "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	for _, name := range []string{"denon", "velociraptor", "trex"} {
		f, err := os.Create(filepath.Join(binDir, name))
		require.NoError(t, err)
		f.Close()
	}

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"deno": {Output: "", Err: nil}, // IsInstalled check
			"deno --version": {
				Output: "deno 1.40.2\nv8 12.1.285.27\ntypescript 5.3.3",
			},
		},
	}

	s := NewDenoScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Deno)

	deno := result.Deno
	assert.Equal(t, "1.40.2", deno.Version)
	assert.Len(t, deno.GlobalPackages, 3)

	names := make([]string, len(deno.GlobalPackages))
	for i, p := range deno.GlobalPackages {
		names[i] = p.Name
		assert.Empty(t, p.Version, "deno global packages should have no version")
	}
	assert.ElementsMatch(t, []string{"denon", "velociraptor", "trex"}, names)
}

func TestDenoScanner_Scan_NoBinDir(t *testing.T) {
	homeDir := t.TempDir()
	// No ~/.deno/bin/ directory created.

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"deno": {Output: "", Err: nil},
			"deno --version": {
				Output: "deno 1.40.2\nv8 12.1.285.27\ntypescript 5.3.3",
			},
		},
	}

	s := NewDenoScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Deno)

	assert.Equal(t, "1.40.2", result.Deno.Version)
	assert.Empty(t, result.Deno.GlobalPackages)
}

func TestDenoScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// "deno" key absent → IsInstalled returns false
		},
	}

	s := NewDenoScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	assert.Nil(t, result.Deno)
}

func TestDenoScanner_Scan_VersionParsing(t *testing.T) {
	homeDir := t.TempDir()

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"deno": {Output: "", Err: nil},
			"deno --version": {
				Output: "deno 2.0.0-rc.1\nv8 13.0.0.0\ntypescript 5.5.0",
			},
		},
	}

	s := NewDenoScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Deno)

	assert.Equal(t, "2.0.0-rc.1", result.Deno.Version)
}

func TestDenoScanner_Scan_BinDirSkipsSubdirs(t *testing.T) {
	homeDir := t.TempDir()

	binDir := filepath.Join(homeDir, ".deno", "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))

	// Create a file and a subdirectory — subdirectory should not appear as a package.
	f, err := os.Create(filepath.Join(binDir, "myscript"))
	require.NoError(t, err)
	f.Close()
	require.NoError(t, os.MkdirAll(filepath.Join(binDir, "subdir"), 0o755))

	mock := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"deno": {Output: "", Err: nil},
			"deno --version": {
				Output: "deno 1.40.2",
			},
		},
	}

	s := NewDenoScanner(homeDir, mock)
	result, err := s.Scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, result.Deno)

	assert.Len(t, result.Deno.GlobalPackages, 1)
	assert.Equal(t, "myscript", result.Deno.GlobalPackages[0].Name)
}
