package editors

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNeovimScanner_Name(t *testing.T) {
	s := NewNeovimScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "neovim", s.Name())
}

func TestNeovimScanner_Description(t *testing.T) {
	s := NewNeovimScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestNeovimScanner_Category(t *testing.T) {
	s := NewNeovimScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "editors", s.Category())
}

func TestNeovimScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()
	// "nvim" key absent → IsInstalled returns false.
	cmd := &util.MockCommandRunner{Responses: map[string]util.MockResponse{}}
	s := NewNeovimScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Neovim)
}

func TestNeovimScanner_Scan_InstalledNoConfig(t *testing.T) {
	homeDir := t.TempDir()
	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"nvim": {Output: "", Err: nil}, // IsInstalled → true
		},
	}
	s := NewNeovimScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Neovim)
	assert.Empty(t, result.Neovim.ConfigDir)
	assert.Empty(t, result.Neovim.PluginManager)
}

func TestNeovimScanner_Scan_WithConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config", "nvim")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"nvim": {Output: "", Err: nil},
		},
	}
	s := NewNeovimScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Neovim)
	assert.Equal(t, filepath.Join(".config", "nvim"), result.Neovim.ConfigDir)
	assert.Empty(t, result.Neovim.PluginManager)
}

func TestNeovimScanner_Scan_PluginManager_Lazy(t *testing.T) {
	homeDir := t.TempDir()
	lazyDir := filepath.Join(homeDir, ".local", "share", "nvim", "lazy")
	require.NoError(t, os.MkdirAll(lazyDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"nvim": {Output: "", Err: nil},
		},
	}
	s := NewNeovimScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Neovim)
	assert.Equal(t, "lazy.nvim", result.Neovim.PluginManager)
}

func TestNeovimScanner_Scan_PluginManager_Packer(t *testing.T) {
	homeDir := t.TempDir()
	packerDir := filepath.Join(homeDir, ".local", "share", "nvim", "site", "pack", "packer")
	require.NoError(t, os.MkdirAll(packerDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"nvim": {Output: "", Err: nil},
		},
	}
	s := NewNeovimScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Neovim)
	assert.Equal(t, "packer", result.Neovim.PluginManager)
}

func TestNeovimScanner_Scan_PluginManager_VimPlug(t *testing.T) {
	homeDir := t.TempDir()
	pluggedDir := filepath.Join(homeDir, ".local", "share", "nvim", "plugged")
	require.NoError(t, os.MkdirAll(pluggedDir, 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"nvim": {Output: "", Err: nil},
		},
	}
	s := NewNeovimScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Neovim)
	assert.Equal(t, "vim-plug", result.Neovim.PluginManager)
}

func TestNeovimScanner_Scan_LazyTakesPriorityOverPacker(t *testing.T) {
	homeDir := t.TempDir()
	// Both lazy and packer dirs exist — lazy.nvim should win (checked first).
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".local", "share", "nvim", "lazy"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".local", "share", "nvim", "site", "pack", "packer"), 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"nvim": {Output: "", Err: nil},
		},
	}
	s := NewNeovimScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Neovim)
	assert.Equal(t, "lazy.nvim", result.Neovim.PluginManager)
}

func TestNeovimScanner_Scan_ConfigDirAndPluginManager(t *testing.T) {
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".config", "nvim")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".local", "share", "nvim", "lazy"), 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"nvim": {Output: "", Err: nil},
		},
	}
	s := NewNeovimScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Neovim)
	assert.Equal(t, filepath.Join(".config", "nvim"), result.Neovim.ConfigDir)
	assert.Equal(t, "lazy.nvim", result.Neovim.PluginManager)
}
