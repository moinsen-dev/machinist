package bundler

import (
	"strings"
	"testing"
	"time"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMeta() domain.Meta {
	return domain.Meta{
		CreatedAt:        time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		SourceHostname:   "test-mac",
		SourceOSVersion:  "darwin",
		SourceArch:       "arm64",
		MachinistVersion: "0.1.0",
	}
}

func TestGenerateRestoreScript_HomebrewOnly(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Taps: []string{"homebrew/core", "homebrew/cask"},
			Formulae: []domain.Package{
				{Name: "git", Version: "2.40"},
				{Name: "jq"},
			},
			Casks: []domain.Package{
				{Name: "firefox"},
			},
			Services: []domain.ServiceEntry{
				{Name: "postgresql", Status: "started"},
				{Name: "redis", Status: "stopped"},
			},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, `brew tap homebrew/core`)
	assert.Contains(t, script, `brew install git`)
	assert.Contains(t, script, `brew install --cask firefox`)
	assert.Contains(t, script, `brew services start postgresql`)
	assert.NotContains(t, script, "Shell Configuration")
}

func TestGenerateRestoreScript_ShellOnly(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
			ConfigFiles: []domain.ConfigFile{
				{Source: ".zshrc", BundlePath: "configs/.zshrc"},
				{Source: ".zprofile", BundlePath: "configs/.zprofile"},
			},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, `cp "configs/`)
	assert.Contains(t, script, `chsh -s "/bin/zsh"`)
	assert.NotContains(t, script, "Homebrew")
}

func TestGenerateRestoreScript_Combined(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{
				{Name: "git"},
			},
		},
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
			ConfigFiles: []domain.ConfigFile{
				{Source: ".zshrc", BundlePath: "configs/.zshrc"},
			},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, `stage "Homebrew"`)
	assert.Contains(t, script, `stage "Shell Configuration"`)
	assert.Contains(t, script, `brew install git`)
	assert.Contains(t, script, `chsh -s "/bin/zsh"`)
}

func TestGenerateRestoreScript_EmptySections(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "LOGFILE=")
	assert.Contains(t, script, "machinist restore completed")
	assert.NotContains(t, script, `stage "Homebrew"`)
	assert.NotContains(t, script, `stage "Shell Configuration"`)
}

func TestGenerateRestoreScript_EmptyHomebrew(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Taps:     []string{},
			Formulae: []domain.Package{},
			Casks:    []domain.Package{},
			Services: []domain.ServiceEntry{},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, `stage "Homebrew"`)
	assert.NotContains(t, script, "brew install ")
	assert.NotContains(t, script, "brew tap ")
	assert.NotContains(t, script, "brew services start")
}

func TestGenerateRestoreScript_IdempotentInstall(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{
				{Name: "wget"},
				{Name: "curl"},
			},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Each formula install should be guarded by a brew list check (idempotent pattern)
	for _, name := range []string{"wget", "curl"} {
		expected := "brew list " + name + " &>/dev/null || brew install " + name
		assert.True(t, strings.Contains(script, expected),
			"expected idempotent install pattern for %s, got:\n%s", name, script)
	}
}

func TestGenerateRestoreScript_ShellNoDefaultShell(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Shell: &domain.ShellSection{
			DefaultShell: "",
			ConfigFiles: []domain.ConfigFile{
				{Source: ".zshrc", BundlePath: "configs/.zshrc"},
				{Source: ".bashrc", BundlePath: "configs/.bashrc"},
			},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Config file restore commands should be present
	assert.Contains(t, script, `cp "configs/.zshrc"`)
	assert.Contains(t, script, `cp "configs/.bashrc"`)
	assert.Contains(t, script, `stage "Shell Configuration"`)

	// chsh must NOT appear when DefaultShell is empty
	assert.NotContains(t, script, "chsh")
}

func TestGenerateRestoreScript_HomebrewServicesNotStarted(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{
				{Name: "postgresql"},
				{Name: "redis"},
			},
			Services: []domain.ServiceEntry{
				{Name: "postgresql", Status: "none"},
				{Name: "redis", Status: "stopped"},
			},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Formulae should still be installed
	assert.Contains(t, script, "brew install postgresql")
	assert.Contains(t, script, "brew install redis")

	// No service should be started because none have Status=="started"
	assert.NotContains(t, script, "brew services start")
}

func TestGenerateRestoreScript_MetaFieldsInScript(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: domain.Meta{
			CreatedAt:        time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
			SourceHostname:   "test-mac",
			SourceOSVersion:  "darwin 15.0",
			SourceArch:       "arm64",
			MachinistVersion: "0.1.0",
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// The header comments should contain the Meta fields
	assert.True(t, strings.Contains(script, "test-mac"),
		"expected hostname in script header, got:\n%s", script)
	assert.True(t, strings.Contains(script, "darwin 15.0"),
		"expected OS version in script header, got:\n%s", script)
	assert.True(t, strings.Contains(script, "arm64"),
		"expected arch in script header, got:\n%s", script)
}
