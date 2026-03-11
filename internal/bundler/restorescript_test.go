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

func TestGenerateRestoreScript_HomebrewPATHInit(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, `/opt/homebrew/bin/brew shellenv`)
	assert.Contains(t, script, `/usr/local/bin/brew shellenv`)
}

func TestGenerateRestoreScript_SSHBeforeGitRepos(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		SSH: &domain.SSHSection{
			Keys: []string{"id_ed25519"},
		},
		GPG: &domain.GPGSection{
			Keys: []string{"ABC123"},
		},
		Git: &domain.GitSection{
			ConfigFiles: []domain.ConfigFile{
				{Source: ".gitconfig", BundlePath: "configs/.gitconfig"},
			},
		},
		GitRepos: &domain.GitReposSection{
			Repositories: []domain.Repository{
				{Remote: "git@github.com:user/repo.git", Path: "~/work/repo"},
			},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	sshIdx := strings.Index(script, `run_stage "SSH Keys"`)
	gpgIdx := strings.Index(script, `run_stage "GPG Keys"`)
	gitCfgIdx := strings.Index(script, `run_stage "Git Configuration"`)
	gitReposIdx := strings.Index(script, `run_stage "Git Repositories"`)

	require.Greater(t, sshIdx, 0, "SSH stage not found")
	require.Greater(t, gpgIdx, 0, "GPG stage not found")
	require.Greater(t, gitCfgIdx, 0, "Git Config stage not found")
	require.Greater(t, gitReposIdx, 0, "Git Repos stage not found")

	assert.Less(t, sshIdx, gitReposIdx, "SSH must come before Git Repos")
	assert.Less(t, gpgIdx, gitReposIdx, "GPG must come before Git Repos")
	assert.Less(t, gitCfgIdx, gitReposIdx, "Git Config must come before Git Repos")
}

func TestGenerateRestoreScript_GPGBundlePath(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		GPG: &domain.GPGSection{
			Encrypted: true,
			Keys:      []string{"ABC123"},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, `configs/gpg/ABC123`)
	assert.NotContains(t, script, `configs/gnupg/`)
}

func TestGenerateRestoreScript_SSHConfigBundlePaths(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		SSH: &domain.SSHSection{
			ConfigFile: ".ssh/config",
			KnownHosts: ".ssh/known_hosts",
			Keys:       []string{"id_ed25519"},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, `if [ -f "configs/ssh/config" ]`)
	assert.Contains(t, script, `cp "configs/ssh/config" "$HOME/.ssh/config"`)
	assert.Contains(t, script, `if [ -f "configs/ssh/known_hosts" ]`)
	assert.Contains(t, script, `cp "configs/ssh/known_hosts" "$HOME/.ssh/known_hosts"`)
}

func TestGenerateRestoreScript_ConfigFileBundlePaths(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Docker: &domain.DockerSection{
			ConfigFile: ".docker/config.json",
		},
		AWS: &domain.AWSSection{
			ConfigFile: ".aws/config",
		},
		Kubernetes: &domain.KubernetesSection{
			ConfigFile: ".kube/config",
		},
		Terraform: &domain.TerraformSection{
			ConfigFile: ".terraformrc",
		},
		Flyio: &domain.FlyioSection{
			ConfigFile: ".fly/config.yml",
		},
		Rectangle: &domain.RectangleSection{
			ConfigFile: "Library/Preferences/com.knollsoft.Rectangle.plist",
		},
		BetterTouchTool: &domain.BetterTouchToolSection{
			ConfigFile: "Library/Application Support/BetterTouchTool/btt_data.json",
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Each template must use configs/<prefix>/basename, not the raw source path
	assert.Contains(t, script, `"configs/docker/config.json"`)
	assert.Contains(t, script, `"configs/aws/config"`)
	assert.Contains(t, script, `"configs/kubernetes/config"`)
	assert.Contains(t, script, `"configs/terraform/.terraformrc"`)
	assert.Contains(t, script, `"configs/flyio/config.yml"`)
	assert.Contains(t, script, `"configs/rectangle/com.knollsoft.Rectangle.plist"`)
	assert.Contains(t, script, `configs/bettertouchtool/btt_data.json`)
}

func TestGenerateRestoreScript_ConfigDirBundlePaths(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		GitHubCLI: &domain.GitHubCLISection{
			ConfigDir: ".config/gh",
		},
		Neovim: &domain.NeovimSection{
			ConfigDir: ".config/nvim",
		},
		Vercel: &domain.VercelSection{
			ConfigDir: ".vercel",
		},
		Firebase: &domain.FirebaseSection{
			ConfigDir: ".config/firebase",
		},
		CloudflareWrangler: &domain.CloudflareSection{
			ConfigDir: ".config/.wrangler",
		},
		Karabiner: &domain.KarabinerSection{
			ConfigDir: ".config/karabiner",
		},
		Alfred: &domain.AlfredSection{
			ConfigDir: "Library/Application Support/Alfred",
		},
		OnePassword: &domain.OnePasswordSection{
			ConfigDir: ".config/op",
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, `"configs/github-cli/"`)
	assert.Contains(t, script, `"configs/neovim/"`)
	assert.Contains(t, script, `"configs/vercel/"`)
	assert.Contains(t, script, `"configs/firebase/"`)
	assert.Contains(t, script, `"configs/cloudflare/"`)
	assert.Contains(t, script, `"configs/karabiner/"`)
	assert.Contains(t, script, `configs/alfred/`)
	assert.Contains(t, script, `"configs/onepassword/"`)
}

func TestGenerateRestoreScript_XDGConfigRestoresToolDirs(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		XDGConfig: &domain.XDGConfigSection{
			AutoDetected: []string{"bat", "lazygit", "starship"},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	assert.Contains(t, script, `configs/xdg-config/bat`)
	assert.Contains(t, script, `cp -R "configs/xdg-config/bat/" "$HOME/.config/bat/"`)
	assert.Contains(t, script, `cp -R "configs/xdg-config/lazygit/" "$HOME/.config/lazygit/"`)
	assert.Contains(t, script, `cp -R "configs/xdg-config/starship/" "$HOME/.config/starship/"`)
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

// ── GenerateRestoreScripts (plural) tests ────────────────────────

func TestGenerateRestoreScripts_ReturnsMapOfScripts(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
		},
		Node: &domain.NodeSection{
			Versions: []string{"20.0.0"},
		},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	// Should contain orchestrator and groups with data
	assert.Contains(t, scripts, "install.command")
	assert.Contains(t, scripts, "01-foundation.sh")
	assert.Contains(t, scripts, "02-shell.sh")
	assert.Contains(t, scripts, "03-runtimes.sh")

	// Should NOT contain groups without data
	assert.NotContains(t, scripts, "04-editors.sh")
	assert.NotContains(t, scripts, "05-infrastructure.sh")
	assert.NotContains(t, scripts, "06-repos.sh")
	assert.NotContains(t, scripts, "07-system.sh")
}

func TestGenerateRestoreScripts_GroupScriptIsStandalone(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	foundation := scripts["01-foundation.sh"]
	require.NotEmpty(t, foundation)

	assert.Contains(t, foundation, "#!/bin/bash")
	assert.Contains(t, foundation, "LOGFILE=")
	assert.Contains(t, foundation, `run_stage "Homebrew"`)
	assert.Contains(t, foundation, "brew install git")
}

func TestGenerateRestoreScripts_OrchestratorRunsGroupsInOrder(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
		Shell: &domain.ShellSection{
			DefaultShell: "/bin/zsh",
		},
		Node: &domain.NodeSection{
			Versions: []string{"20.0.0"},
		},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	orch := scripts["install.command"]
	require.NotEmpty(t, orch)

	// All three scripts mentioned in order
	idx1 := strings.Index(orch, "01-foundation.sh")
	idx2 := strings.Index(orch, "02-shell.sh")
	idx3 := strings.Index(orch, "03-runtimes.sh")

	require.Greater(t, idx1, 0, "01-foundation.sh not found in orchestrator")
	require.Greater(t, idx2, 0, "02-shell.sh not found in orchestrator")
	require.Greater(t, idx3, 0, "03-runtimes.sh not found in orchestrator")

	assert.Less(t, idx1, idx2, "01-foundation must come before 02-shell")
	assert.Less(t, idx2, idx3, "02-shell must come before 03-runtimes")
}

func TestGenerateRestoreScripts_EmptySnapshot(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	// Only the orchestrator should be returned
	assert.Len(t, scripts, 1)
	assert.Contains(t, scripts, "install.command")
}

func TestGenerateRestoreScripts_GroupStageCount(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
		SSH: &domain.SSHSection{
			Keys: []string{"id_ed25519"},
		},
		Git: &domain.GitSection{
			ConfigFiles: []domain.ConfigFile{
				{Source: ".gitconfig", BundlePath: "configs/.gitconfig"},
			},
		},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	foundation := scripts["01-foundation.sh"]
	require.NotEmpty(t, foundation)

	// Foundation has Homebrew + SSH + Git = 3 stages (GPG is nil)
	assert.Contains(t, foundation, "STAGE_TOTAL=3")
}
