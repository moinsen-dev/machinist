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

	assert.Contains(t, script, `run_stage "Homebrew"`)
	assert.Contains(t, script, `run_stage "Shell Configuration"`)
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
	assert.NotContains(t, script, `run_stage "Homebrew"`)
	assert.NotContains(t, script, `run_stage "Shell Configuration"`)
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

	assert.Contains(t, script, `run_stage "Homebrew"`)
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
	assert.Contains(t, script, `run_stage "Shell Configuration"`)

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

// Phase 6 tests

func TestGenerateRestoreScript_HasProgressCounter(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Script should contain stage counting variables
	assert.Contains(t, script, "STAGE_NUM=0")
	assert.Contains(t, script, "STAGE_TOTAL=")
	assert.Contains(t, script, "STAGE_PASS=0")
	assert.Contains(t, script, "STAGE_FAIL=0")

	// The stage function should include the counter format
	assert.Contains(t, script, "[$STAGE_NUM/$STAGE_TOTAL]")

	// STAGE_TOTAL should reflect the actual number of stages (2: Homebrew + Shell)
	assert.Contains(t, script, "STAGE_TOTAL=2")
}

func TestGenerateRestoreScript_HasArchCheck(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Architecture check should be present
	assert.Contains(t, script, "uname -m")
	assert.Contains(t, script, "CURRENT_ARCH=")
	assert.Contains(t, script, "SOURCE_ARCH=")
	assert.Contains(t, script, "Architecture mismatch")
}

func TestGenerateRestoreScript_HasErrorRecovery(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Script should NOT use set -e (which would abort on first error)
	assert.NotContains(t, script, "set -e")
	assert.NotContains(t, script, "set -euo")

	// Should use set -uo pipefail instead
	assert.Contains(t, script, "set -uo pipefail")

	// Should have run_stage function for error recovery
	assert.Contains(t, script, "run_stage()")
	assert.Contains(t, script, "run_stage \"Homebrew\"")

	// Should track failures
	assert.Contains(t, script, "STAGE_FAIL=$((STAGE_FAIL + 1))")
	assert.Contains(t, script, "STAGE_PASS=$((STAGE_PASS + 1))")

	// Summary at the end
	assert.Contains(t, script, "machinist restore completed in")
	assert.Contains(t, script, "stages succeeded")
	assert.Contains(t, script, "stages failed")
}

func TestGenerateRestoreScript_HasTiming(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Should capture start and end time
	assert.Contains(t, script, "START_TIME=$(date +%s)")
	assert.Contains(t, script, "END_TIME=$(date +%s)")
	assert.Contains(t, script, "ELAPSED=$((END_TIME - START_TIME))")
}

func TestGenerateRestoreScript_StageCountMatchesSections(t *testing.T) {
	// 0 sections = 0 stages
	snap0 := &domain.Snapshot{Meta: newMeta()}
	script0, err := GenerateRestoreScript(snap0)
	require.NoError(t, err)
	assert.Contains(t, script0, "STAGE_TOTAL=0")

	// 3 sections = 3 stages
	snap3 := &domain.Snapshot{
		Meta:     newMeta(),
		Homebrew: &domain.HomebrewSection{},
		Shell:    &domain.ShellSection{},
		Folders:  &domain.FoldersSection{},
	}
	script3, err := GenerateRestoreScript(snap3)
	require.NoError(t, err)
	assert.Contains(t, script3, "STAGE_TOTAL=3")
}
