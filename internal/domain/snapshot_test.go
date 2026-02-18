package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSnapshot_PopulatesMeta(t *testing.T) {
	before := time.Now()
	snap := NewSnapshot("Moinsen-MacBook-Pro", "15.3", "arm64", "0.1.0")
	after := time.Now()

	require.NotNil(t, snap)
	assert.Equal(t, "Moinsen-MacBook-Pro", snap.Meta.SourceHostname)
	assert.Equal(t, "15.3", snap.Meta.SourceOSVersion)
	assert.Equal(t, "arm64", snap.Meta.SourceArch)
	assert.Equal(t, "0.1.0", snap.Meta.MachinistVersion)
	assert.False(t, snap.Meta.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.True(t, !snap.Meta.CreatedAt.Before(before), "CreatedAt should be >= test start time")
	assert.True(t, !snap.Meta.CreatedAt.After(after), "CreatedAt should be <= test end time")
	assert.Equal(t, float64(0), snap.Meta.ScanDurationSecs)
}

func TestNewSnapshot_AllSectionsNil(t *testing.T) {
	snap := NewSnapshot("host", "15.3", "arm64", "0.1.0")

	assert.Nil(t, snap.Homebrew, "Homebrew should be nil by default")
	assert.Nil(t, snap.Shell, "Shell should be nil by default")
	assert.Nil(t, snap.Terminal, "Terminal should be nil by default")
	assert.Nil(t, snap.Tmux, "Tmux should be nil by default")
	assert.Nil(t, snap.Git, "Git should be nil by default")
	assert.Nil(t, snap.GitHubCLI, "GitHubCLI should be nil by default")
	assert.Nil(t, snap.GitRepos, "GitRepos should be nil by default")
	assert.Nil(t, snap.Node, "Node should be nil by default")
	assert.Nil(t, snap.Python, "Python should be nil by default")
	assert.Nil(t, snap.Rust, "Rust should be nil by default")
	assert.Nil(t, snap.Java, "Java should be nil by default")
	assert.Nil(t, snap.Flutter, "Flutter should be nil by default")
	assert.Nil(t, snap.Go, "Go should be nil by default")
	assert.Nil(t, snap.Asdf, "Asdf should be nil by default")
	assert.Nil(t, snap.VSCode, "VSCode should be nil by default")
	assert.Nil(t, snap.Cursor, "Cursor should be nil by default")
	assert.Nil(t, snap.Neovim, "Neovim should be nil by default")
	assert.Nil(t, snap.JetBrains, "JetBrains should be nil by default")
	assert.Nil(t, snap.Xcode, "Xcode should be nil by default")
	assert.Nil(t, snap.Docker, "Docker should be nil by default")
	assert.Nil(t, snap.AWS, "AWS should be nil by default")
	assert.Nil(t, snap.Kubernetes, "Kubernetes should be nil by default")
	assert.Nil(t, snap.Terraform, "Terraform should be nil by default")
	assert.Nil(t, snap.Vercel, "Vercel should be nil by default")
	assert.Nil(t, snap.MacOSDefaults, "MacOSDefaults should be nil by default")
	assert.Nil(t, snap.Locale, "Locale should be nil by default")
	assert.Nil(t, snap.LoginItems, "LoginItems should be nil by default")
	assert.Nil(t, snap.HostsFile, "HostsFile should be nil by default")
	assert.Nil(t, snap.Apps, "Apps should be nil by default")
	assert.Nil(t, snap.Raycast, "Raycast should be nil by default")
	assert.Nil(t, snap.Karabiner, "Karabiner should be nil by default")
	assert.Nil(t, snap.Rectangle, "Rectangle should be nil by default")
	assert.Nil(t, snap.SSH, "SSH should be nil by default")
	assert.Nil(t, snap.GPG, "GPG should be nil by default")
	assert.Nil(t, snap.XDGConfig, "XDGConfig should be nil by default")
	assert.Nil(t, snap.Folders, "Folders should be nil by default")
	assert.Nil(t, snap.Fonts, "Fonts should be nil by default")
	assert.Nil(t, snap.EnvFiles, "EnvFiles should be nil by default")
	assert.Nil(t, snap.Crontab, "Crontab should be nil by default")
	assert.Nil(t, snap.LaunchAgents, "LaunchAgents should be nil by default")
	assert.Nil(t, snap.Network, "Network should be nil by default")
	assert.Nil(t, snap.Browser, "Browser should be nil by default")
	assert.Nil(t, snap.AITools, "AITools should be nil by default")
	assert.Nil(t, snap.APITools, "APITools should be nil by default")
	assert.Nil(t, snap.Databases, "Databases should be nil by default")
	assert.Nil(t, snap.Registries, "Registries should be nil by default")
}

func TestSnapshot_WithPopulatedHomebrew(t *testing.T) {
	snap := NewSnapshot("host", "15.3", "arm64", "0.1.0")
	snap.Homebrew = &HomebrewSection{
		Taps:     []string{"homebrew/core", "homebrew/cask"},
		Formulae: []Package{{Name: "git", Version: "2.43.0"}},
		Casks:    []Package{{Name: "visual-studio-code"}},
		Services: []ServiceEntry{{Name: "postgresql@16", Status: "started"}},
	}

	require.NotNil(t, snap.Homebrew)
	assert.Len(t, snap.Homebrew.Taps, 2)
	assert.Len(t, snap.Homebrew.Formulae, 1)
	assert.Equal(t, "git", snap.Homebrew.Formulae[0].Name)
	assert.Len(t, snap.Homebrew.Casks, 1)
	assert.Len(t, snap.Homebrew.Services, 1)
	assert.Equal(t, "started", snap.Homebrew.Services[0].Status)
}

func TestSnapshot_MultipleSectionsIndependently(t *testing.T) {
	snap := NewSnapshot("host", "15.3", "arm64", "0.1.0")

	// Populate only Shell
	snap.Shell = &ShellSection{
		DefaultShell: "/bin/zsh",
		Framework:    "oh-my-zsh",
		Prompt:       "starship",
	}

	// Populate only Git
	snap.Git = &GitSection{
		SigningMethod:    "ssh",
		CredentialHelper: "osxkeychain",
	}

	// Populate only Docker
	snap.Docker = &DockerSection{
		Runtime:              "docker-desktop",
		FrequentlyUsedImages: []string{"postgres:16-alpine", "redis:7-alpine"},
	}

	require.NotNil(t, snap.Shell)
	require.NotNil(t, snap.Git)
	require.NotNil(t, snap.Docker)

	assert.Equal(t, "/bin/zsh", snap.Shell.DefaultShell)
	assert.Equal(t, "oh-my-zsh", snap.Shell.Framework)
	assert.Equal(t, "starship", snap.Shell.Prompt)

	assert.Equal(t, "ssh", snap.Git.SigningMethod)
	assert.Equal(t, "osxkeychain", snap.Git.CredentialHelper)

	assert.Equal(t, "docker-desktop", snap.Docker.Runtime)
	assert.Len(t, snap.Docker.FrequentlyUsedImages, 2)

	// Other sections remain nil
	assert.Nil(t, snap.Homebrew)
	assert.Nil(t, snap.Node)
	assert.Nil(t, snap.VSCode)
}

func TestSnapshot_ShellSection_FullFields(t *testing.T) {
	shell := &ShellSection{
		DefaultShell: "/bin/zsh",
		Framework:    "oh-my-zsh",
		Prompt:       "starship",
		ConfigFiles: []ConfigFile{
			{Source: "~/.zshrc", BundlePath: "configs/shell/.zshrc"},
			{Source: "~/.zshenv", BundlePath: "configs/shell/.zshenv"},
		},
		OhMyZshCustomPlugins: []string{"zsh-autosuggestions", "zsh-syntax-highlighting"},
	}

	assert.Equal(t, "/bin/zsh", shell.DefaultShell)
	assert.Len(t, shell.ConfigFiles, 2)
	assert.Len(t, shell.OhMyZshCustomPlugins, 2)
}

func TestSnapshot_NodeSection(t *testing.T) {
	node := &NodeSection{
		Manager:        "fnm",
		Versions:       []string{"20.11.0", "18.19.0"},
		DefaultVersion: "20.11.0",
		GlobalPackages: []Package{
			{Name: "typescript", Version: "5.3.3"},
			{Name: "pnpm", Version: "8.14.0"},
		},
	}

	assert.Equal(t, "fnm", node.Manager)
	assert.Len(t, node.Versions, 2)
	assert.Equal(t, "20.11.0", node.DefaultVersion)
	assert.Len(t, node.GlobalPackages, 2)
}

func TestSnapshot_MacOSDefaultsSection(t *testing.T) {
	defaults := &MacOSDefaultsSection{
		Dock: &DockConfig{AutoHide: true, TileSize: 48, Orientation: "bottom", MinimizeEffect: "scale"},
		Finder: &FinderConfig{ShowHidden: true, ShowPathBar: true, ShowStatusBar: true, DefaultView: "list"},
		Keyboard: &KeyboardConfig{KeyRepeat: 2, InitialKeyRepeat: 15},
		Trackpad: &TrackpadConfig{TapToClick: true, TrackingSpeed: 3.0},
	}

	require.NotNil(t, defaults.Dock)
	assert.True(t, defaults.Dock.AutoHide)
	assert.Equal(t, 48, defaults.Dock.TileSize)

	require.NotNil(t, defaults.Finder)
	assert.True(t, defaults.Finder.ShowHidden)

	require.NotNil(t, defaults.Keyboard)
	assert.Equal(t, 2, defaults.Keyboard.KeyRepeat)

	require.NotNil(t, defaults.Trackpad)
	assert.True(t, defaults.Trackpad.TapToClick)
	assert.Equal(t, 3.0, defaults.Trackpad.TrackingSpeed)
}

func TestSnapshot_SSHSection(t *testing.T) {
	ssh := &SSHSection{
		Encrypted:  true,
		ConfigFile: "configs/ssh/config",
		Keys:       []string{"configs/ssh/id_ed25519.age", "configs/ssh/id_ed25519.pub"},
		KnownHosts: "configs/ssh/known_hosts",
	}

	assert.True(t, ssh.Encrypted)
	assert.Equal(t, "configs/ssh/config", ssh.ConfigFile)
	assert.Len(t, ssh.Keys, 2)
	assert.Equal(t, "configs/ssh/known_hosts", ssh.KnownHosts)
}

func TestSnapshot_NetworkSection(t *testing.T) {
	net := &NetworkSection{
		PreferredWifi: []string{"HomeNetwork", "Office-5G"},
		DNS:           &DNSConfig{Interface: "Wi-Fi", Servers: []string{"1.1.1.1", "8.8.8.8"}},
		VPNConfigs: []ConfigFile{
			{Source: "~/.config/wireguard/wg0.conf", BundlePath: "configs/network/wg0.conf", Encrypted: true},
		},
	}

	assert.Len(t, net.PreferredWifi, 2)
	require.NotNil(t, net.DNS)
	assert.Equal(t, "Wi-Fi", net.DNS.Interface)
	assert.Len(t, net.VPNConfigs, 1)
	assert.True(t, net.VPNConfigs[0].Encrypted)
}

func TestSnapshot_GitReposSection(t *testing.T) {
	repos := &GitReposSection{
		SearchPaths: []string{"~/Code", "~/Projects"},
		Repositories: []Repository{
			{Path: "~/Code/machinist", Remote: "git@github.com:moinsen/machinist.git", Branch: "main", Shallow: false},
			{Path: "~/Code/huge-monorepo", Remote: "git@github.com:company/monorepo.git", Branch: "main", Shallow: true},
		},
	}

	assert.Len(t, repos.SearchPaths, 2)
	assert.Len(t, repos.Repositories, 2)
	assert.False(t, repos.Repositories[0].Shallow)
	assert.True(t, repos.Repositories[1].Shallow)
}

func TestSnapshot_RustSection(t *testing.T) {
	rust := &RustSection{
		Toolchains:       []string{"stable", "nightly"},
		DefaultToolchain: "stable",
		Components:       []string{"clippy", "rustfmt", "rust-analyzer"},
		CargoPackages: []Package{
			{Name: "cargo-watch", Version: "8.5.2"},
		},
	}

	assert.Len(t, rust.Toolchains, 2)
	assert.Equal(t, "stable", rust.DefaultToolchain)
	assert.Len(t, rust.Components, 3)
	assert.Len(t, rust.CargoPackages, 1)
}

func TestSnapshot_AIToolsSection(t *testing.T) {
	ai := &AIToolsSection{
		ClaudeCodeConfig: "configs/ai/claude/",
		OllamaModels:     []string{"llama3:8b", "codellama:13b"},
	}

	assert.Equal(t, "configs/ai/claude/", ai.ClaudeCodeConfig)
	assert.Len(t, ai.OllamaModels, 2)
}

func TestSnapshot_AppsSection(t *testing.T) {
	apps := &AppsSection{
		AppStore: []InstalledApp{
			{Name: "Xcode", ID: 497799835},
			{Name: "Magnet", ID: 441258766},
		},
	}

	assert.Len(t, apps.AppStore, 2)
	assert.Equal(t, "Xcode", apps.AppStore[0].Name)
	assert.Equal(t, 497799835, apps.AppStore[0].ID)
}
