package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/moinsen-dev/machinist/internal/scanner"
	"github.com/moinsen-dev/machinist/internal/scanner/packages"
	"github.com/moinsen-dev/machinist/internal/scanner/shell"
	"github.com/moinsen-dev/machinist/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullPipeline(t *testing.T) {
	// 1. Set up MockCommandRunner with realistic responses.
	mockCmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			// IsInstalled("brew") checks for key "brew" with nil error.
			"brew": {Output: "", Err: nil},
			"brew list --formula --versions": {
				Output: "git 2.43.0\nnode 21.5.0\ngo 1.21.5",
			},
			"brew list --cask": {
				Output: "firefox\nvisual-studio-code\ndocker",
			},
			"brew tap": {
				Output: "homebrew/core\nhomebrew/cask\nhomebrew/services",
			},
			"brew services list": {
				Output: "Name       Status  User    File\npostgresql started doedel ~/Library/LaunchAgents/...\nredis      none",
			},
			"sh -c echo $SHELL": {
				Output: "/bin/zsh",
			},
		},
	}

	// 2. Create temp directory as homeDir with shell config files.
	homeDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".zshrc"), []byte(`export PATH="/usr/local/bin:$PATH"`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".zshenv"), []byte(`export EDITOR=vim`), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".oh-my-zsh"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".config"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".config", "starship.toml"), []byte("[character]\nsymbol = \"→\""), 0644))

	// 3. Create Registry and register both scanners.
	reg := scanner.NewRegistry()
	require.NoError(t, reg.Register(packages.NewHomebrewScanner(mockCmd)))
	require.NoError(t, reg.Register(shell.NewShellConfigScanner(homeDir, mockCmd)))

	// 4. Run ScanAll.
	snap, errs := reg.ScanAll(context.Background())
	require.Empty(t, errs)
	require.NotNil(t, snap)

	// 5. Assert snapshot has all expected data.

	// Homebrew section
	require.NotNil(t, snap.Homebrew, "Homebrew section should be populated")
	assert.Len(t, snap.Homebrew.Formulae, 3)
	assert.Equal(t, "git", snap.Homebrew.Formulae[0].Name)
	assert.Equal(t, "2.43.0", snap.Homebrew.Formulae[0].Version)
	assert.Equal(t, "node", snap.Homebrew.Formulae[1].Name)
	assert.Equal(t, "21.5.0", snap.Homebrew.Formulae[1].Version)
	assert.Equal(t, "go", snap.Homebrew.Formulae[2].Name)
	assert.Equal(t, "1.21.5", snap.Homebrew.Formulae[2].Version)
	assert.Len(t, snap.Homebrew.Casks, 3)
	assert.Equal(t, "firefox", snap.Homebrew.Casks[0].Name)
	assert.Equal(t, "visual-studio-code", snap.Homebrew.Casks[1].Name)
	assert.Equal(t, "docker", snap.Homebrew.Casks[2].Name)
	assert.Len(t, snap.Homebrew.Taps, 3)
	assert.Equal(t, []string{"homebrew/core", "homebrew/cask", "homebrew/services"}, snap.Homebrew.Taps)
	assert.Len(t, snap.Homebrew.Services, 2)
	assert.Equal(t, "postgresql", snap.Homebrew.Services[0].Name)
	assert.Equal(t, "started", snap.Homebrew.Services[0].Status)
	assert.Equal(t, "redis", snap.Homebrew.Services[1].Name)
	assert.Equal(t, "none", snap.Homebrew.Services[1].Status)

	// Shell section
	require.NotNil(t, snap.Shell, "Shell section should be populated")
	assert.Equal(t, "/bin/zsh", snap.Shell.DefaultShell)
	assert.Equal(t, "oh-my-zsh", snap.Shell.Framework)
	assert.Equal(t, "starship", snap.Shell.Prompt)
	assert.Len(t, snap.Shell.ConfigFiles, 3)

	// Check config file sources (order: known configs first, then starship).
	sources := make([]string, len(snap.Shell.ConfigFiles))
	for i, cf := range snap.Shell.ConfigFiles {
		sources[i] = cf.Source
	}
	assert.Contains(t, sources, "~/.zshrc")
	assert.Contains(t, sources, "~/.zshenv")
	assert.Contains(t, sources, "~/.config/starship.toml")

	// All config files should have a content hash.
	for _, cf := range snap.Shell.ConfigFiles {
		assert.NotEmpty(t, cf.ContentHash, "ContentHash should be set for %s", cf.Source)
	}

	// 6. Marshal to TOML.
	tomlBytes, err := domain.MarshalManifest(snap)
	require.NoError(t, err)
	tomlStr := string(tomlBytes)

	assert.Contains(t, tomlStr, "[homebrew]")
	assert.Contains(t, tomlStr, "[shell]")
	assert.Contains(t, tomlStr, "\"git\"")
	assert.Contains(t, tomlStr, "\"firefox\"")
	assert.Contains(t, tomlStr, "/bin/zsh")

	// 7. Unmarshal back and verify roundtrip.
	snap2, err := domain.UnmarshalManifest(tomlBytes)
	require.NoError(t, err)
	require.NotNil(t, snap2)

	// Homebrew roundtrip.
	require.NotNil(t, snap2.Homebrew)
	assert.Equal(t, snap.Homebrew.Formulae, snap2.Homebrew.Formulae)
	assert.Equal(t, snap.Homebrew.Casks, snap2.Homebrew.Casks)
	assert.Equal(t, snap.Homebrew.Taps, snap2.Homebrew.Taps)
	assert.Equal(t, snap.Homebrew.Services, snap2.Homebrew.Services)

	// Shell roundtrip.
	require.NotNil(t, snap2.Shell)
	assert.Equal(t, snap.Shell.DefaultShell, snap2.Shell.DefaultShell)
	assert.Equal(t, snap.Shell.Framework, snap2.Shell.Framework)
	assert.Equal(t, snap.Shell.Prompt, snap2.Shell.Prompt)
	assert.Equal(t, snap.Shell.ConfigFiles, snap2.Shell.ConfigFiles)
}

func TestFullPipeline_NoBrewInstalled(t *testing.T) {
	// MockCommandRunner with NO "brew" entry so IsInstalled returns false.
	mockCmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{
			"sh -c echo $SHELL": {
				Output: "/bin/zsh",
			},
		},
	}

	homeDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(homeDir, ".zshrc"), []byte(`export PATH="/usr/local/bin:$PATH"`), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(homeDir, ".oh-my-zsh"), 0755))

	reg := scanner.NewRegistry()
	require.NoError(t, reg.Register(packages.NewHomebrewScanner(mockCmd)))
	require.NoError(t, reg.Register(shell.NewShellConfigScanner(homeDir, mockCmd)))

	snap, errs := reg.ScanAll(context.Background())
	require.Empty(t, errs)
	require.NotNil(t, snap)

	// Homebrew section should be nil since brew is not installed.
	assert.Nil(t, snap.Homebrew, "Homebrew section should be nil when brew is not installed")

	// Shell section should still be populated.
	require.NotNil(t, snap.Shell)
	assert.Equal(t, "/bin/zsh", snap.Shell.DefaultShell)
	assert.Len(t, snap.Shell.ConfigFiles, 1)
	assert.Equal(t, "oh-my-zsh", snap.Shell.Framework)

	// Marshal and verify no [homebrew] in TOML.
	tomlBytes, err := domain.MarshalManifest(snap)
	require.NoError(t, err)
	tomlStr := string(tomlBytes)

	assert.NotContains(t, tomlStr, "[homebrew]")
	assert.Contains(t, tomlStr, "[shell]")
}

func TestFullPipeline_EmptySystem(t *testing.T) {
	// MockCommandRunner with no responses at all.
	mockCmd := &util.MockCommandRunner{
		Responses: map[string]util.MockResponse{},
	}

	// Empty temp dir -- no config files.
	homeDir := t.TempDir()

	reg := scanner.NewRegistry()
	require.NoError(t, reg.Register(packages.NewHomebrewScanner(mockCmd)))
	require.NoError(t, reg.Register(shell.NewShellConfigScanner(homeDir, mockCmd)))

	snap, errs := reg.ScanAll(context.Background())
	require.NotNil(t, snap)

	// The shell scanner's Run for "sh -c echo $SHELL" will fail (unknown command),
	// but it doesn't return an error -- it just leaves DefaultShell empty.
	// The homebrew scanner returns empty result (brew not installed).
	// So we expect no fatal errors from ScanAll.
	assert.Empty(t, errs)

	// Homebrew should be nil (brew not installed).
	assert.Nil(t, snap.Homebrew)

	// Shell section should exist but be mostly empty.
	require.NotNil(t, snap.Shell)
	assert.Empty(t, snap.Shell.DefaultShell)
	assert.Empty(t, snap.Shell.ConfigFiles)
	assert.Empty(t, snap.Shell.Framework)
	assert.Empty(t, snap.Shell.Prompt)

	// Marshal to TOML -- should produce minimal output with just [meta].
	tomlBytes, err := domain.MarshalManifest(snap)
	require.NoError(t, err)
	tomlStr := string(tomlBytes)

	assert.Contains(t, tomlStr, "[meta]")
	assert.NotContains(t, tomlStr, "[homebrew]")

	// The shell section will appear in TOML because DefaultShell has the `toml:"default_shell"`
	// tag without omitempty, so the empty struct will be serialized.
	// Verify roundtrip still works.
	snap2, err := domain.UnmarshalManifest(tomlBytes)
	require.NoError(t, err)
	require.NotNil(t, snap2)
	assert.Nil(t, snap2.Homebrew)

	// Verify TOML doesn't contain any brew-related content.
	for _, line := range strings.Split(tomlStr, "\n") {
		assert.NotContains(t, strings.ToLower(line), "formulae")
	}
}
