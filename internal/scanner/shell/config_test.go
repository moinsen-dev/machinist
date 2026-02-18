package shell

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellConfigScanner_Name(t *testing.T) {
	s := NewShellConfigScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "shell", s.Name())
}

func TestShellConfigScanner_Description(t *testing.T) {
	s := NewShellConfigScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.NotEmpty(t, s.Description())
}

func TestShellConfigScanner_Category(t *testing.T) {
	s := NewShellConfigScanner("/tmp", &util.MockCommandRunner{Responses: map[string]util.MockResponse{}})
	assert.Equal(t, "shell", s.Category())
}

func TestShellConfigScanner_Scan_ZshFiles(t *testing.T) {
	homeDir := t.TempDir()

	// Create zsh config files with sample content.
	files := map[string]string{
		".zshrc":    "# zshrc config\nexport PATH=$PATH:/usr/local/bin",
		".zshenv":   "# zshenv\nexport EDITOR=vim",
		".zprofile": "# zprofile\neval \"$(brew shellenv)\"",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(homeDir, name), []byte(content), 0644))
	}

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/zsh"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)

	assert.Len(t, result.Shell.ConfigFiles, 3)
	for _, cf := range result.Shell.ConfigFiles {
		assert.NotEmpty(t, cf.Source, "Source should not be empty")
		assert.NotEmpty(t, cf.ContentHash, "ContentHash should not be empty")
	}

	// Verify specific files are present.
	sources := make(map[string]bool)
	for _, cf := range result.Shell.ConfigFiles {
		sources[cf.Source] = true
	}
	assert.True(t, sources["~/.zshrc"])
	assert.True(t, sources["~/.zshenv"])
	assert.True(t, sources["~/.zprofile"])
}

func TestShellConfigScanner_Scan_BashFiles(t *testing.T) {
	homeDir := t.TempDir()

	files := map[string]string{
		".bashrc":       "# bashrc\nexport PS1='$ '",
		".bash_profile": "# bash_profile\nsource ~/.bashrc",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(homeDir, name), []byte(content), 0644))
	}

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/bash"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)

	assert.Len(t, result.Shell.ConfigFiles, 2)

	sources := make(map[string]bool)
	for _, cf := range result.Shell.ConfigFiles {
		sources[cf.Source] = true
	}
	assert.True(t, sources["~/.bashrc"])
	assert.True(t, sources["~/.bash_profile"])
}

func TestShellConfigScanner_Scan_DefaultShell(t *testing.T) {
	homeDir := t.TempDir()

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/zsh"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Equal(t, "/bin/zsh", result.Shell.DefaultShell)
}

func TestShellConfigScanner_Scan_OhMyZsh(t *testing.T) {
	homeDir := t.TempDir()

	// Create .oh-my-zsh directory.
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".oh-my-zsh"), 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/zsh"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Equal(t, "oh-my-zsh", result.Shell.Framework)
}

func TestShellConfigScanner_Scan_Starship(t *testing.T) {
	homeDir := t.TempDir()

	// Create .config/starship.toml.
	configDir := filepath.Join(homeDir, ".config")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "starship.toml"), []byte("[character]\nsuccess_symbol = \"➜\""), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/zsh"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Equal(t, "starship", result.Shell.Prompt)

	// Verify starship.toml is in ConfigFiles.
	found := false
	for _, cf := range result.Shell.ConfigFiles {
		if cf.Source == "~/.config/starship.toml" {
			found = true
			assert.NotEmpty(t, cf.ContentHash)
			break
		}
	}
	assert.True(t, found, "starship.toml should be in ConfigFiles")
}

func TestShellConfigScanner_Scan_P10k(t *testing.T) {
	homeDir := t.TempDir()

	// Create .p10k.zsh.
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".p10k.zsh"), []byte("# p10k config\ntypeset -g POWERLEVEL9K_MODE=nerdfont"), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/zsh"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Equal(t, "p10k", result.Shell.Prompt)

	// Verify .p10k.zsh is in ConfigFiles.
	found := false
	for _, cf := range result.Shell.ConfigFiles {
		if cf.Source == "~/.p10k.zsh" {
			found = true
			assert.NotEmpty(t, cf.ContentHash)
			break
		}
	}
	assert.True(t, found, ".p10k.zsh should be in ConfigFiles")
}

func TestShellConfigScanner_Scan_EmptyHome(t *testing.T) {
	homeDir := t.TempDir()

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: ""},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Empty(t, result.Shell.ConfigFiles)
	assert.Empty(t, result.Shell.Framework)
	assert.Empty(t, result.Shell.Prompt)
}

func TestShellConfigScanner_Scan_Combined(t *testing.T) {
	homeDir := t.TempDir()

	// Create .zshrc.
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".zshrc"), []byte("# zshrc"), 0644))

	// Create .oh-my-zsh directory.
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".oh-my-zsh"), 0755))

	// Create .config/starship.toml.
	configDir := filepath.Join(homeDir, ".config")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "starship.toml"), []byte("[character]\nsuccess_symbol = \"➜\""), 0644))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/zsh"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)

	// Check default shell.
	assert.Equal(t, "/bin/zsh", result.Shell.DefaultShell)

	// Check framework.
	assert.Equal(t, "oh-my-zsh", result.Shell.Framework)

	// Check prompt.
	assert.Equal(t, "starship", result.Shell.Prompt)

	// Check config files: .zshrc + starship.toml = 2.
	assert.Len(t, result.Shell.ConfigFiles, 2)

	sources := make(map[string]bool)
	for _, cf := range result.Shell.ConfigFiles {
		sources[cf.Source] = true
	}
	assert.True(t, sources["~/.zshrc"])
	assert.True(t, sources["~/.config/starship.toml"])
}

func TestShellConfigScanner_Scan_Prezto(t *testing.T) {
	homeDir := t.TempDir()

	// Create .zprezto directory.
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".zprezto"), 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/zsh"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Equal(t, "prezto", result.Shell.Framework)
}

func TestShellConfigScanner_Scan_OhMyBash(t *testing.T) {
	homeDir := t.TempDir()

	// Create .oh-my-bash directory.
	require.NoError(t, os.Mkdir(filepath.Join(homeDir, ".oh-my-bash"), 0755))

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/bash"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Equal(t, "oh-my-bash", result.Shell.Framework)
}

func TestShellConfigScanner_Scan_AllDotfiles(t *testing.T) {
	homeDir := t.TempDir()

	// Create all 11 known shell config files.
	allFiles := []string{
		".zshrc", ".zshenv", ".zprofile", ".zlogin", ".zlogout",
		".bashrc", ".bash_profile", ".bash_login", ".bash_logout",
		".profile", ".inputrc",
	}
	for _, name := range allFiles {
		require.NoError(t, os.WriteFile(filepath.Join(homeDir, name), []byte("# "+name), 0644))
	}

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "/bin/zsh"},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Len(t, result.Shell.ConfigFiles, 11)

	// Verify all expected sources are present.
	sources := make(map[string]bool)
	for _, cf := range result.Shell.ConfigFiles {
		sources[cf.Source] = true
		assert.NotEmpty(t, cf.ContentHash, "ContentHash should not be empty for %s", cf.Source)
	}
	for _, name := range allFiles {
		assert.True(t, sources["~/"+name], "expected %s to be in ConfigFiles", name)
	}
}

func TestShellConfigScanner_Scan_ShellDetectionError(t *testing.T) {
	homeDir := t.TempDir()

	cmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {Output: "", Err: fmt.Errorf("command not found")},
		},
	}
	s := NewShellConfigScanner(homeDir, cmd)
	result, err := s.Scan(context.Background())

	// Scan should not return an error even when shell detection fails.
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Shell)
	assert.Empty(t, result.Shell.DefaultShell, "DefaultShell should be empty when shell detection command fails")
}
