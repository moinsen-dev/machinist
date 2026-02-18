// Package profiles provides embedded preset TOML manifests that describe
// typical developer setups. Each profile is a complete Snapshot that can
// be used directly or merged with user overrides.
package profiles

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/moinsen-dev/machinist/internal/domain"
)

//go:embed *.toml
var profileFS embed.FS

// List returns names of all available profiles (without .toml extension), sorted.
func List() ([]string, error) {
	entries, err := profileFS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("reading embedded profiles: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".toml") {
			names = append(names, strings.TrimSuffix(name, ".toml"))
		}
	}
	sort.Strings(names)
	return names, nil
}

// Get returns the Snapshot for a named profile.
func Get(name string) (*domain.Snapshot, error) {
	data, err := profileFS.ReadFile(name + ".toml")
	if err != nil {
		return nil, fmt.Errorf("profile %q not found: %w", name, err)
	}
	snap, err := domain.UnmarshalManifest(data)
	if err != nil {
		return nil, fmt.Errorf("parsing profile %q: %w", name, err)
	}
	return snap, nil
}

// Merge merges a base profile snapshot with overrides. The base provides
// default values; override sections take precedence. For Homebrew, lists
// (taps, formulae, casks) are merged with deduplication. For other pointer
// sections, the override replaces the base if non-nil.
func Merge(base, override *domain.Snapshot) *domain.Snapshot {
	merged := *base // shallow copy

	// Merge Homebrew specially: combine lists.
	if override.Homebrew != nil {
		if merged.Homebrew == nil {
			merged.Homebrew = override.Homebrew
		} else {
			hw := *merged.Homebrew
			hw.Taps = mergeStrings(hw.Taps, override.Homebrew.Taps)
			hw.Formulae = mergePackages(hw.Formulae, override.Homebrew.Formulae)
			hw.Casks = mergePackages(hw.Casks, override.Homebrew.Casks)
			hw.Services = mergeServices(hw.Services, override.Homebrew.Services)
			merged.Homebrew = &hw
		}
	}

	// For all other pointer sections: override wins if non-nil.
	if override.Node != nil {
		merged.Node = override.Node
	}
	if override.Python != nil {
		merged.Python = override.Python
	}
	if override.Rust != nil {
		merged.Rust = override.Rust
	}
	if override.Java != nil {
		merged.Java = override.Java
	}
	if override.Flutter != nil {
		merged.Flutter = override.Flutter
	}
	if override.Go != nil {
		merged.Go = override.Go
	}
	if override.Asdf != nil {
		merged.Asdf = override.Asdf
	}
	if override.Shell != nil {
		merged.Shell = override.Shell
	}
	if override.Terminal != nil {
		merged.Terminal = override.Terminal
	}
	if override.Tmux != nil {
		merged.Tmux = override.Tmux
	}
	if override.Git != nil {
		merged.Git = override.Git
	}
	if override.GitHubCLI != nil {
		merged.GitHubCLI = override.GitHubCLI
	}
	if override.GitRepos != nil {
		merged.GitRepos = override.GitRepos
	}
	if override.VSCode != nil {
		merged.VSCode = override.VSCode
	}
	if override.Cursor != nil {
		merged.Cursor = override.Cursor
	}
	if override.Neovim != nil {
		merged.Neovim = override.Neovim
	}
	if override.JetBrains != nil {
		merged.JetBrains = override.JetBrains
	}
	if override.Xcode != nil {
		merged.Xcode = override.Xcode
	}
	if override.Docker != nil {
		merged.Docker = override.Docker
	}
	if override.AWS != nil {
		merged.AWS = override.AWS
	}
	if override.Kubernetes != nil {
		merged.Kubernetes = override.Kubernetes
	}
	if override.Terraform != nil {
		merged.Terraform = override.Terraform
	}
	if override.Vercel != nil {
		merged.Vercel = override.Vercel
	}
	if override.MacOSDefaults != nil {
		merged.MacOSDefaults = override.MacOSDefaults
	}
	if override.Locale != nil {
		merged.Locale = override.Locale
	}
	if override.LoginItems != nil {
		merged.LoginItems = override.LoginItems
	}
	if override.HostsFile != nil {
		merged.HostsFile = override.HostsFile
	}
	if override.Apps != nil {
		merged.Apps = override.Apps
	}
	if override.Raycast != nil {
		merged.Raycast = override.Raycast
	}
	if override.Karabiner != nil {
		merged.Karabiner = override.Karabiner
	}
	if override.Rectangle != nil {
		merged.Rectangle = override.Rectangle
	}
	if override.SSH != nil {
		merged.SSH = override.SSH
	}
	if override.GPG != nil {
		merged.GPG = override.GPG
	}
	if override.XDGConfig != nil {
		merged.XDGConfig = override.XDGConfig
	}
	if override.Folders != nil {
		merged.Folders = override.Folders
	}
	if override.Fonts != nil {
		merged.Fonts = override.Fonts
	}
	if override.EnvFiles != nil {
		merged.EnvFiles = override.EnvFiles
	}
	if override.Crontab != nil {
		merged.Crontab = override.Crontab
	}
	if override.LaunchAgents != nil {
		merged.LaunchAgents = override.LaunchAgents
	}
	if override.Network != nil {
		merged.Network = override.Network
	}
	if override.Browser != nil {
		merged.Browser = override.Browser
	}
	if override.AITools != nil {
		merged.AITools = override.AITools
	}
	if override.APITools != nil {
		merged.APITools = override.APITools
	}
	if override.Databases != nil {
		merged.Databases = override.Databases
	}
	if override.Registries != nil {
		merged.Registries = override.Registries
	}

	return &merged
}

// mergeStrings combines two string slices, deduplicating by value.
func mergeStrings(base, extra []string) []string {
	seen := make(map[string]bool, len(base))
	result := make([]string, 0, len(base)+len(extra))
	for _, s := range base {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range extra {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// mergePackages combines two Package slices, deduplicating by Name.
func mergePackages(base, extra []domain.Package) []domain.Package {
	seen := make(map[string]bool, len(base))
	result := make([]domain.Package, 0, len(base)+len(extra))
	for _, p := range base {
		if !seen[p.Name] {
			seen[p.Name] = true
			result = append(result, p)
		}
	}
	for _, p := range extra {
		if !seen[p.Name] {
			seen[p.Name] = true
			result = append(result, p)
		}
	}
	return result
}

// mergeServices combines two ServiceEntry slices, deduplicating by Name.
func mergeServices(base, extra []domain.ServiceEntry) []domain.ServiceEntry {
	seen := make(map[string]bool, len(base))
	result := make([]domain.ServiceEntry, 0, len(base)+len(extra))
	for _, s := range base {
		if !seen[s.Name] {
			seen[s.Name] = true
			result = append(result, s)
		}
	}
	for _, s := range extra {
		if !seen[s.Name] {
			seen[s.Name] = true
			result = append(result, s)
		}
	}
	return result
}
