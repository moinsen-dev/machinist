package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTimestamp returns a deterministic timestamp for tests.
func testTimestamp() time.Time {
	return time.Date(2026, 2, 18, 14, 30, 0, 0, time.UTC)
}

// testMeta returns a deterministic Meta for tests.
func testMeta() Meta {
	return Meta{
		CreatedAt:        testTimestamp(),
		SourceHostname:   "test-mac",
		SourceOSVersion:  "15.3",
		SourceArch:       "arm64",
		MachinistVersion: "0.1.0",
		ScanDurationSecs: 1.23,
	}
}

func TestMarshalManifest_HomebrewOnly(t *testing.T) {
	s := &Snapshot{
		Meta: testMeta(),
		Homebrew: &HomebrewSection{
			Taps: []string{"homebrew/core", "homebrew/cask"},
			Formulae: []Package{
				{Name: "git", Version: "2.44.0"},
				{Name: "go", Version: "1.22.1"},
			},
			Casks: []Package{
				{Name: "visual-studio-code"},
				{Name: "iterm2"},
			},
			Services: []ServiceEntry{
				{Name: "postgresql@16", Status: "started"},
			},
		},
	}

	data, err := MarshalManifest(s)
	require.NoError(t, err)

	tomlStr := string(data)

	// Verify homebrew section and its sub-sections are present
	assert.Contains(t, tomlStr, "[homebrew]")
	assert.Contains(t, tomlStr, "[[homebrew.formulae]]")
	assert.Contains(t, tomlStr, "[[homebrew.casks]]")
	assert.Contains(t, tomlStr, "[[homebrew.services]]")

	// Verify tap names
	assert.Contains(t, tomlStr, "homebrew/core")
	assert.Contains(t, tomlStr, "homebrew/cask")

	// Verify formula names and versions
	assert.Contains(t, tomlStr, `name = "git"`)
	assert.Contains(t, tomlStr, `version = "2.44.0"`)
	assert.Contains(t, tomlStr, `name = "go"`)

	// Verify cask names
	assert.Contains(t, tomlStr, `name = "visual-studio-code"`)
	assert.Contains(t, tomlStr, `name = "iterm2"`)

	// Verify service entries
	assert.Contains(t, tomlStr, `name = "postgresql@16"`)
	assert.Contains(t, tomlStr, `status = "started"`)

	// Verify other sections are NOT present
	assert.NotContains(t, tomlStr, "[shell]")
	assert.NotContains(t, tomlStr, "[git]")
	assert.NotContains(t, tomlStr, "[docker]")
}

func TestMarshalManifest_ShellOnly(t *testing.T) {
	s := &Snapshot{
		Meta: testMeta(),
		Shell: &ShellSection{
			DefaultShell: "/bin/zsh",
			Framework:    "oh-my-zsh",
			ConfigFiles: []ConfigFile{
				{Source: "~/.zshrc", BundlePath: "shell/zshrc"},
				{Source: "~/.zprofile", BundlePath: "shell/zprofile"},
			},
		},
	}

	data, err := MarshalManifest(s)
	require.NoError(t, err)

	tomlStr := string(data)

	assert.Contains(t, tomlStr, "[shell]")
	assert.Contains(t, tomlStr, `default_shell = "/bin/zsh"`)
	assert.Contains(t, tomlStr, `framework = "oh-my-zsh"`)
	assert.Contains(t, tomlStr, "[[shell.config_files]]")
	assert.Contains(t, tomlStr, `source = "~/.zshrc"`)
	assert.Contains(t, tomlStr, `bundle_path = "shell/zshrc"`)

	// Verify other sections are NOT present
	assert.NotContains(t, tomlStr, "[homebrew]")
	assert.NotContains(t, tomlStr, "[docker]")
}

func TestMarshalManifest_EmptySections(t *testing.T) {
	s := &Snapshot{
		Meta: testMeta(),
	}

	data, err := MarshalManifest(s)
	require.NoError(t, err)

	tomlStr := string(data)

	// Only meta section should be present
	assert.Contains(t, tomlStr, "[meta]")

	// No other top-level sections should appear
	sectionNames := []string{
		"[homebrew]", "[node]", "[python]", "[rust]", "[java]",
		"[flutter]", "[go]", "[asdf]", "[shell]", "[terminal]",
		"[tmux]", "[git]", "[github_cli]", "[git_repos]", "[vscode]",
		"[cursor]", "[neovim]", "[jetbrains]", "[xcode]", "[docker]",
		"[aws]", "[kubernetes]", "[terraform]", "[vercel]",
		"[macos_defaults]", "[locale]", "[login_items]", "[hosts_file]",
		"[apps]", "[raycast]", "[karabiner]", "[rectangle]",
		"[ssh]", "[gpg]", "[xdg_config]", "[folders]", "[fonts]",
		"[env_files]", "[crontab]", "[launchagents]", "[network]",
		"[browser]", "[ai_tools]", "[api_tools]", "[databases]",
		"[registries]",
	}
	for _, section := range sectionNames {
		assert.NotContains(t, tomlStr, section, "nil section %s should not appear in TOML output", section)
	}
}

func TestUnmarshalManifest(t *testing.T) {
	tomlStr := `
[meta]
created_at = 2026-02-18T14:30:00Z
source_hostname = "test-mac"
source_os_version = "15.3"
source_arch = "arm64"
machinist_version = "0.1.0"
scan_duration_secs = 1.23

[homebrew]
taps = ["homebrew/core", "homebrew/cask"]

[[homebrew.formulae]]
name = "git"
version = "2.44.0"

[[homebrew.casks]]
name = "iterm2"

[[homebrew.services]]
name = "postgresql@16"
status = "started"

[shell]
default_shell = "/bin/zsh"
framework = "oh-my-zsh"

[[shell.config_files]]
source = "~/.zshrc"
bundle_path = "shell/zshrc"
`

	s, err := UnmarshalManifest([]byte(tomlStr))
	require.NoError(t, err)

	// Verify meta
	assert.Equal(t, "test-mac", s.Meta.SourceHostname)
	assert.Equal(t, "15.3", s.Meta.SourceOSVersion)
	assert.Equal(t, "arm64", s.Meta.SourceArch)
	assert.Equal(t, "0.1.0", s.Meta.MachinistVersion)
	assert.InDelta(t, 1.23, s.Meta.ScanDurationSecs, 0.001)
	assert.Equal(t, testTimestamp(), s.Meta.CreatedAt)

	// Verify homebrew
	require.NotNil(t, s.Homebrew)
	assert.Equal(t, []string{"homebrew/core", "homebrew/cask"}, s.Homebrew.Taps)
	require.Len(t, s.Homebrew.Formulae, 1)
	assert.Equal(t, "git", s.Homebrew.Formulae[0].Name)
	assert.Equal(t, "2.44.0", s.Homebrew.Formulae[0].Version)
	require.Len(t, s.Homebrew.Casks, 1)
	assert.Equal(t, "iterm2", s.Homebrew.Casks[0].Name)
	require.Len(t, s.Homebrew.Services, 1)
	assert.Equal(t, "postgresql@16", s.Homebrew.Services[0].Name)
	assert.Equal(t, "started", s.Homebrew.Services[0].Status)

	// Verify shell
	require.NotNil(t, s.Shell)
	assert.Equal(t, "/bin/zsh", s.Shell.DefaultShell)
	assert.Equal(t, "oh-my-zsh", s.Shell.Framework)
	require.Len(t, s.Shell.ConfigFiles, 1)
	assert.Equal(t, "~/.zshrc", s.Shell.ConfigFiles[0].Source)
	assert.Equal(t, "shell/zshrc", s.Shell.ConfigFiles[0].BundlePath)

	// Verify nil sections remain nil
	assert.Nil(t, s.Docker)
	assert.Nil(t, s.Git)
	assert.Nil(t, s.Node)
}

func TestManifestRoundtrip(t *testing.T) {
	original := &Snapshot{
		Meta: testMeta(),
		Homebrew: &HomebrewSection{
			Taps: []string{"homebrew/core"},
			Formulae: []Package{
				{Name: "git", Version: "2.44.0"},
				{Name: "ripgrep", Version: "14.1.0"},
			},
			Casks: []Package{
				{Name: "firefox"},
			},
			Services: []ServiceEntry{
				{Name: "redis", Status: "started"},
			},
		},
		Shell: &ShellSection{
			DefaultShell: "/bin/zsh",
			Framework:    "oh-my-zsh",
			Prompt:       "starship",
			ConfigFiles: []ConfigFile{
				{Source: "~/.zshrc", BundlePath: "shell/zshrc"},
			},
		},
	}

	// Marshal
	data, err := MarshalManifest(original)
	require.NoError(t, err)

	// Unmarshal
	restored, err := UnmarshalManifest(data)
	require.NoError(t, err)

	// Compare meta
	assert.Equal(t, original.Meta.SourceHostname, restored.Meta.SourceHostname)
	assert.Equal(t, original.Meta.SourceOSVersion, restored.Meta.SourceOSVersion)
	assert.Equal(t, original.Meta.SourceArch, restored.Meta.SourceArch)
	assert.Equal(t, original.Meta.MachinistVersion, restored.Meta.MachinistVersion)
	assert.InDelta(t, original.Meta.ScanDurationSecs, restored.Meta.ScanDurationSecs, 0.001)

	// Compare homebrew
	require.NotNil(t, restored.Homebrew)
	assert.Equal(t, original.Homebrew.Taps, restored.Homebrew.Taps)
	assert.Equal(t, original.Homebrew.Formulae, restored.Homebrew.Formulae)
	assert.Equal(t, original.Homebrew.Casks, restored.Homebrew.Casks)
	assert.Equal(t, original.Homebrew.Services, restored.Homebrew.Services)

	// Compare shell
	require.NotNil(t, restored.Shell)
	assert.Equal(t, original.Shell.DefaultShell, restored.Shell.DefaultShell)
	assert.Equal(t, original.Shell.Framework, restored.Shell.Framework)
	assert.Equal(t, original.Shell.Prompt, restored.Shell.Prompt)
	assert.Equal(t, original.Shell.ConfigFiles, restored.Shell.ConfigFiles)

	// Nil sections should remain nil
	assert.Nil(t, restored.Docker)
	assert.Nil(t, restored.Git)
}

func TestWriteAndReadManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.toml")

	original := &Snapshot{
		Meta: testMeta(),
		Homebrew: &HomebrewSection{
			Taps:     []string{"homebrew/core"},
			Formulae: []Package{{Name: "wget", Version: "1.24"}},
		},
		Shell: &ShellSection{
			DefaultShell: "/bin/zsh",
			Framework:    "oh-my-zsh",
			ConfigFiles: []ConfigFile{
				{Source: "~/.zshrc", BundlePath: "shell/zshrc"},
			},
		},
	}

	// Write to file
	err := WriteManifest(original, manifestPath)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(manifestPath)
	require.NoError(t, err)

	// Read back
	restored, err := ReadManifest(manifestPath)
	require.NoError(t, err)

	// Compare fields
	assert.Equal(t, original.Meta.SourceHostname, restored.Meta.SourceHostname)
	assert.Equal(t, original.Meta.MachinistVersion, restored.Meta.MachinistVersion)

	require.NotNil(t, restored.Homebrew)
	assert.Equal(t, original.Homebrew.Taps, restored.Homebrew.Taps)
	assert.Equal(t, original.Homebrew.Formulae, restored.Homebrew.Formulae)

	require.NotNil(t, restored.Shell)
	assert.Equal(t, original.Shell.DefaultShell, restored.Shell.DefaultShell)
	assert.Equal(t, original.Shell.Framework, restored.Shell.Framework)
	assert.Equal(t, original.Shell.ConfigFiles, restored.Shell.ConfigFiles)
}

func TestMarshalManifest_MultipleSections(t *testing.T) {
	s := &Snapshot{
		Meta: testMeta(),
		Homebrew: &HomebrewSection{
			Taps:     []string{"homebrew/core"},
			Formulae: []Package{{Name: "git", Version: "2.44.0"}},
		},
		Shell: &ShellSection{
			DefaultShell: "/bin/zsh",
			Framework:    "oh-my-zsh",
		},
		Git: &GitSection{
			SigningMethod:    "gpg",
			CredentialHelper: "osxkeychain",
			ConfigFiles: []ConfigFile{
				{Source: "~/.gitconfig", BundlePath: "git/gitconfig"},
			},
		},
		Docker: &DockerSection{
			Runtime:              "docker-desktop",
			FrequentlyUsedImages: []string{"postgres:16", "redis:7"},
			ConfigFile:           "~/.docker/config.json",
		},
	}

	data, err := MarshalManifest(s)
	require.NoError(t, err)

	tomlStr := string(data)

	// All four sections should be present
	assert.Contains(t, tomlStr, "[meta]")
	assert.Contains(t, tomlStr, "[homebrew]")
	assert.Contains(t, tomlStr, "[shell]")
	assert.Contains(t, tomlStr, "[git]")
	assert.Contains(t, tomlStr, "[docker]")

	// Spot-check content from each section
	assert.Contains(t, tomlStr, `name = "git"`)
	assert.Contains(t, tomlStr, `default_shell = "/bin/zsh"`)
	assert.Contains(t, tomlStr, `signing_method = "gpg"`)
	assert.Contains(t, tomlStr, `runtime = "docker-desktop"`)

	// Verify sections that are nil are NOT present
	nilSections := []string{"[node]", "[python]", "[rust]", "[java]", "[vscode]"}
	for _, section := range nilSections {
		// Use a more precise check: the section header at line start or after newline
		lines := strings.Split(tomlStr, "\n")
		found := false
		for _, line := range lines {
			if strings.TrimSpace(line) == section {
				found = true
				break
			}
		}
		assert.False(t, found, "nil section %s should not appear in TOML output", section)
	}
}

func TestUnmarshalManifest_InvalidTOML(t *testing.T) {
	_, err := UnmarshalManifest([]byte("{{{"))
	assert.Error(t, err, "invalid TOML should produce an error")
}

func TestReadManifest_FileNotFound(t *testing.T) {
	_, err := ReadManifest("/tmp/nonexistent_path_abc123/manifest.toml")
	assert.Error(t, err, "reading a nonexistent file should produce an error")
}

func TestWriteManifest_BadPath(t *testing.T) {
	s := &Snapshot{Meta: testMeta()}
	err := WriteManifest(s, "/nonexistent_dir/foo/bar/manifest.toml")
	assert.Error(t, err, "writing to an invalid path should produce an error")
}

func TestMarshalManifest_NilSnapshot(t *testing.T) {
	s := &Snapshot{}
	data, err := MarshalManifest(s)
	require.NoError(t, err, "marshalling an empty Snapshot should not error")
	assert.Contains(t, string(data), "[meta]", "output should contain [meta] section")
}
