package shell

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTmuxScanner_Name(t *testing.T) {
	s := NewTmuxScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "tmux", s.Name())
}

func TestTmuxScanner_Description(t *testing.T) {
	s := NewTmuxScanner("/tmp", &util.MockCommandRunner{})
	assert.NotEmpty(t, s.Description())
}

func TestTmuxScanner_Category(t *testing.T) {
	s := NewTmuxScanner("/tmp", &util.MockCommandRunner{})
	assert.Equal(t, "shell", s.Category())
}

func TestTmuxScanner_Scan_NotInstalled(t *testing.T) {
	homeDir := t.TempDir()

	// No "tmux" key in responses → IsInstalled returns false.
	cmd := &util.MockCommandRunner{Responses: map[string]util.MockResponse{}}

	s := NewTmuxScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.Tmux, "Tmux section should be nil when tmux is not installed")
}

func TestTmuxScanner_Scan_NoConfigFiles(t *testing.T) {
	homeDir := t.TempDir()

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"tmux": {Output: "3.3a", Err: nil}, // IsInstalled
		},
	}

	s := NewTmuxScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Tmux)
	assert.Empty(t, result.Tmux.ConfigFiles)
	assert.Empty(t, result.Tmux.TPMPlugins)
}

func TestTmuxScanner_Scan_HomeTmuxConf(t *testing.T) {
	homeDir := t.TempDir()

	confContent := `# tmux config
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-sensible'
set -g @plugin "tmux-plugins/tmux-resurrect"
`
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".tmux.conf"), []byte(confContent), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"tmux": {Output: "3.3a", Err: nil},
		},
	}

	s := NewTmuxScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Tmux)

	// Config file recorded.
	require.Len(t, result.Tmux.ConfigFiles, 1)
	assert.Equal(t, ".tmux.conf", result.Tmux.ConfigFiles[0].Source)
	assert.NotEmpty(t, result.Tmux.ConfigFiles[0].ContentHash)

	// Plugins extracted.
	assert.Equal(t, []string{
		"tmux-plugins/tpm",
		"tmux-plugins/tmux-sensible",
		"tmux-plugins/tmux-resurrect",
	}, result.Tmux.TPMPlugins)
}

func TestTmuxScanner_Scan_XDGConfigTmuxConf(t *testing.T) {
	homeDir := t.TempDir()

	xdgDir := filepath.Join(homeDir, ".config", "tmux")
	require.NoError(t, os.MkdirAll(xdgDir, 0755))

	confContent := "set -g @plugin 'tmux-plugins/tpm'\n"
	require.NoError(t, os.WriteFile(filepath.Join(xdgDir, "tmux.conf"), []byte(confContent), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"tmux": {Output: "3.3a", Err: nil},
		},
	}

	s := NewTmuxScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Tmux)

	require.Len(t, result.Tmux.ConfigFiles, 1)
	assert.Equal(t, ".config/tmux/tmux.conf", result.Tmux.ConfigFiles[0].Source)

	assert.Equal(t, []string{"tmux-plugins/tpm"}, result.Tmux.TPMPlugins)
}

func TestTmuxScanner_Scan_BothConfigFiles(t *testing.T) {
	homeDir := t.TempDir()

	// ~/.tmux.conf
	require.NoError(t, os.WriteFile(
		filepath.Join(homeDir, ".tmux.conf"),
		[]byte("set -g @plugin 'tmux-plugins/tpm'\n"),
		0644,
	))

	// ~/.config/tmux/tmux.conf
	xdgDir := filepath.Join(homeDir, ".config", "tmux")
	require.NoError(t, os.MkdirAll(xdgDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(xdgDir, "tmux.conf"),
		[]byte("set -g @plugin 'tmux-plugins/tmux-sensible'\n"),
		0644,
	))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"tmux": {Output: "3.3a", Err: nil},
		},
	}

	s := NewTmuxScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Tmux)

	// Both config files captured.
	require.Len(t, result.Tmux.ConfigFiles, 2)

	sources := make(map[string]bool)
	for _, cf := range result.Tmux.ConfigFiles {
		sources[cf.Source] = true
	}
	assert.True(t, sources[".tmux.conf"])
	assert.True(t, sources[".config/tmux/tmux.conf"])

	// Plugins from both files, deduplicated.
	assert.Equal(t, []string{
		"tmux-plugins/tpm",
		"tmux-plugins/tmux-sensible",
	}, result.Tmux.TPMPlugins)
}

func TestTmuxScanner_Scan_DuplicatePluginsDeduped(t *testing.T) {
	homeDir := t.TempDir()

	confContent := `set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-sensible'
`
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".tmux.conf"), []byte(confContent), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"tmux": {Output: "3.3a", Err: nil},
		},
	}

	s := NewTmuxScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result.Tmux)

	// Duplicate 'tpm' should appear only once.
	assert.Equal(t, []string{
		"tmux-plugins/tpm",
		"tmux-plugins/tmux-sensible",
	}, result.Tmux.TPMPlugins)
}
