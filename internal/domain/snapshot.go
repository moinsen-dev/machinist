package domain

import "time"

// Meta contains metadata about the snapshot: when, where, and how it was created.
type Meta struct {
	CreatedAt        time.Time `toml:"created_at"`
	SourceHostname   string    `toml:"source_hostname"`
	SourceOSVersion  string    `toml:"source_os_version"`
	SourceArch       string    `toml:"source_arch"`
	MachinistVersion string    `toml:"machinist_version"`
	ScanDurationSecs float64   `toml:"scan_duration_secs"`
}

// Snapshot is the root aggregate representing a complete picture of a machine's
// developer environment. Each section is a pointer: nil means "not scanned".
type Snapshot struct {
	Meta          Meta                  `toml:"meta"`
	Homebrew      *HomebrewSection      `toml:"homebrew,omitempty"`
	Node          *NodeSection          `toml:"node,omitempty"`
	Python        *PythonSection        `toml:"python,omitempty"`
	Rust          *RustSection          `toml:"rust,omitempty"`
	Java          *JavaSection          `toml:"java,omitempty"`
	Flutter       *FlutterSection       `toml:"flutter,omitempty"`
	Go            *GoSection            `toml:"go,omitempty"`
	Asdf          *AsdfSection          `toml:"asdf,omitempty"`
	Shell         *ShellSection         `toml:"shell,omitempty"`
	Terminal      *TerminalSection      `toml:"terminal,omitempty"`
	Tmux          *TmuxSection          `toml:"tmux,omitempty"`
	Git           *GitSection           `toml:"git,omitempty"`
	GitHubCLI     *GitHubCLISection     `toml:"github_cli,omitempty"`
	GitRepos      *GitReposSection      `toml:"git_repos,omitempty"`
	VSCode        *VSCodeSection        `toml:"vscode,omitempty"`
	Cursor        *CursorSection        `toml:"cursor,omitempty"`
	Neovim        *NeovimSection        `toml:"neovim,omitempty"`
	JetBrains     *JetBrainsSection     `toml:"jetbrains,omitempty"`
	Xcode         *XcodeSection         `toml:"xcode,omitempty"`
	Docker        *DockerSection        `toml:"docker,omitempty"`
	AWS           *AWSSection           `toml:"aws,omitempty"`
	Kubernetes    *KubernetesSection    `toml:"kubernetes,omitempty"`
	Terraform     *TerraformSection     `toml:"terraform,omitempty"`
	Vercel        *VercelSection        `toml:"vercel,omitempty"`
	MacOSDefaults *MacOSDefaultsSection `toml:"macos_defaults,omitempty"`
	Locale        *LocaleSection        `toml:"locale,omitempty"`
	LoginItems    *LoginItemsSection    `toml:"login_items,omitempty"`
	HostsFile     *HostsFileSection     `toml:"hosts_file,omitempty"`
	Apps          *AppsSection          `toml:"apps,omitempty"`
	Raycast       *RaycastSection       `toml:"raycast,omitempty"`
	Karabiner     *KarabinerSection     `toml:"karabiner,omitempty"`
	Rectangle     *RectangleSection     `toml:"rectangle,omitempty"`
	SSH           *SSHSection           `toml:"ssh,omitempty"`
	GPG           *GPGSection           `toml:"gpg,omitempty"`
	XDGConfig     *XDGConfigSection     `toml:"xdg_config,omitempty"`
	Folders       *FoldersSection       `toml:"folders,omitempty"`
	Fonts         *FontsSection         `toml:"fonts,omitempty"`
	EnvFiles      *EnvFilesSection      `toml:"env_files,omitempty"`
	Crontab       *CrontabSection       `toml:"crontab,omitempty"`
	LaunchAgents  *LaunchAgentsSection  `toml:"launchagents,omitempty"`
	Network       *NetworkSection       `toml:"network,omitempty"`
	Browser       *BrowserSection       `toml:"browser,omitempty"`
	AITools       *AIToolsSection       `toml:"ai_tools,omitempty"`
	APITools      *APIToolsSection      `toml:"api_tools,omitempty"`
	Databases          *DatabasesSection          `toml:"databases,omitempty"`
	Registries         *RegistriesSection         `toml:"registries,omitempty"`
	Deno               *DenoSection               `toml:"deno,omitempty"`
	Bun                *BunSection                `toml:"bun,omitempty"`
	Ruby               *RubySection               `toml:"ruby,omitempty"`
	GCP                *GCPSection                `toml:"gcp,omitempty"`
	Azure              *AzureSection              `toml:"azure,omitempty"`
	Flyio              *FlyioSection              `toml:"flyio,omitempty"`
	Firebase           *FirebaseSection           `toml:"firebase,omitempty"`
	CloudflareWrangler *CloudflareSection         `toml:"cloudflare,omitempty"`
	Alfred             *AlfredSection             `toml:"alfred,omitempty"`
	BetterTouchTool    *BetterTouchToolSection    `toml:"bettertouchtool,omitempty"`
	OnePassword        *OnePasswordSection        `toml:"onepassword,omitempty"`
}

// NewSnapshot creates a new Snapshot with populated Meta fields and all sections set to nil.
func NewSnapshot(hostname, osVersion, arch, machinistVersion string) *Snapshot {
	return &Snapshot{
		Meta: Meta{
			CreatedAt:        time.Now(),
			SourceHostname:   hostname,
			SourceOSVersion:  osVersion,
			SourceArch:       arch,
			MachinistVersion: machinistVersion,
		},
	}
}

// StageCount returns the number of restore stages that will be executed,
// based on which snapshot sections are non-nil.
func (s *Snapshot) StageCount() int {
	count := 0
	if s.Homebrew != nil {
		count++
	}
	if s.Shell != nil {
		count++
	}
	if s.GitRepos != nil {
		count++
	}
	if s.Node != nil {
		count++
	}
	if s.Python != nil {
		count++
	}
	if s.Rust != nil {
		count++
	}
	if s.Java != nil {
		count++
	}
	if s.Flutter != nil {
		count++
	}
	if s.Go != nil {
		count++
	}
	if s.Asdf != nil {
		count++
	}
	if s.Deno != nil {
		count++
	}
	if s.Bun != nil {
		count++
	}
	if s.Ruby != nil {
		count++
	}
	if s.Terminal != nil {
		count++
	}
	if s.Tmux != nil {
		count++
	}
	if s.Git != nil {
		count++
	}
	if s.GitHubCLI != nil {
		count++
	}
	if s.VSCode != nil {
		count++
	}
	if s.Cursor != nil {
		count++
	}
	if s.Neovim != nil {
		count++
	}
	if s.JetBrains != nil {
		count++
	}
	if s.Xcode != nil {
		count++
	}
	if s.Docker != nil {
		count++
	}
	if s.AWS != nil {
		count++
	}
	if s.Kubernetes != nil {
		count++
	}
	if s.Terraform != nil {
		count++
	}
	if s.Vercel != nil {
		count++
	}
	if s.GCP != nil {
		count++
	}
	if s.Azure != nil {
		count++
	}
	if s.Flyio != nil {
		count++
	}
	if s.Firebase != nil {
		count++
	}
	if s.CloudflareWrangler != nil {
		count++
	}
	if s.MacOSDefaults != nil {
		count++
	}
	if s.Locale != nil {
		count++
	}
	if s.LoginItems != nil {
		count++
	}
	if s.HostsFile != nil {
		count++
	}
	if s.Apps != nil {
		count++
	}
	if s.Raycast != nil {
		count++
	}
	if s.Alfred != nil {
		count++
	}
	if s.Karabiner != nil {
		count++
	}
	if s.Rectangle != nil {
		count++
	}
	if s.BetterTouchTool != nil {
		count++
	}
	if s.OnePassword != nil {
		count++
	}
	if s.SSH != nil {
		count++
	}
	if s.GPG != nil {
		count++
	}
	if s.XDGConfig != nil {
		count++
	}
	if s.EnvFiles != nil {
		count++
	}
	if s.Folders != nil {
		count++
	}
	if s.Fonts != nil {
		count++
	}
	if s.Crontab != nil || s.LaunchAgents != nil {
		count++
	}
	if s.Network != nil {
		count++
	}
	if s.Browser != nil {
		count++
	}
	if s.AITools != nil {
		count++
	}
	if s.APITools != nil {
		count++
	}
	if s.Databases != nil {
		count++
	}
	if s.Registries != nil {
		count++
	}
	return count
}

// ---------------------------------------------------------------------------
// Section types — each corresponds to a TOML section in the manifest.
// ---------------------------------------------------------------------------

// HomebrewSection captures Homebrew taps, formulae, casks, and services.
type HomebrewSection struct {
	Taps     []string       `toml:"taps,omitempty"`
	Formulae []Package      `toml:"formulae,omitempty"`
	Casks    []Package      `toml:"casks,omitempty"`
	Services []ServiceEntry `toml:"services,omitempty"`
}

// NodeSection captures Node.js version manager, versions, and global packages.
type NodeSection struct {
	Manager        string    `toml:"manager,omitempty"`
	Versions       []string  `toml:"versions,omitempty"`
	DefaultVersion string    `toml:"default_version,omitempty"`
	GlobalPackages []Package `toml:"global_packages,omitempty"`
}

// PythonSection captures Python version manager, versions, and global packages.
type PythonSection struct {
	Manager        string    `toml:"manager,omitempty"`
	Versions       []string  `toml:"versions,omitempty"`
	DefaultVersion string    `toml:"default_version,omitempty"`
	GlobalPackages []Package `toml:"global_packages,omitempty"`
}

// RustSection captures Rust toolchains, components, and cargo-installed packages.
type RustSection struct {
	Toolchains       []string  `toml:"toolchains,omitempty"`
	DefaultToolchain string    `toml:"default_toolchain,omitempty"`
	Components       []string  `toml:"components,omitempty"`
	CargoPackages    []Package `toml:"cargo_packages,omitempty"`
}

// JavaSection captures Java/Kotlin SDK manager, versions, and JAVA_HOME.
type JavaSection struct {
	Manager        string   `toml:"manager,omitempty"`
	Versions       []string `toml:"versions,omitempty"`
	DefaultVersion string   `toml:"default_version,omitempty"`
	JavaHome       string   `toml:"java_home,omitempty"`
}

// FlutterSection captures Flutter SDK channel, version, and Dart global packages.
type FlutterSection struct {
	Channel            string   `toml:"channel,omitempty"`
	Version            string   `toml:"version,omitempty"`
	DartGlobalPackages []string `toml:"dart_global_packages,omitempty"`
}

// GoSection captures Go version and globally installed packages.
type GoSection struct {
	Version        string    `toml:"version,omitempty"`
	GlobalPackages []Package `toml:"global_packages,omitempty"`
}

// AsdfSection captures asdf/mise plugins, versions, and the tool-versions file.
type AsdfSection struct {
	Manager          string       `toml:"manager,omitempty"` // "asdf" or "mise"
	Plugins          []AsdfPlugin `toml:"plugins,omitempty"`
	ToolVersionsFile string       `toml:"tool_versions_file,omitempty"`
}

// ShellSection captures shell configuration: default shell, framework, prompt, and config files.
type ShellSection struct {
	DefaultShell         string       `toml:"default_shell"`
	Framework            string       `toml:"framework,omitempty"`
	Prompt               string       `toml:"prompt,omitempty"`
	ConfigFiles          []ConfigFile `toml:"config_files,omitempty"`
	OhMyZshCustomPlugins []string     `toml:"oh_my_zsh_custom_plugins,omitempty"`
}

// TerminalSection captures terminal emulator configuration.
type TerminalSection struct {
	App         string       `toml:"app,omitempty"`
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
}

// TmuxSection captures tmux configuration and TPM plugins.
type TmuxSection struct {
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
	TPMPlugins  []string     `toml:"tpm_plugins,omitempty"`
}

// GitSection captures git configuration, signing method, templates, and credential helper.
type GitSection struct {
	ConfigFiles      []ConfigFile `toml:"config_files,omitempty"`
	SigningMethod    string       `toml:"signing_method,omitempty"`
	TemplateDir      string       `toml:"template_dir,omitempty"`
	CredentialHelper string       `toml:"credential_helper,omitempty"`
}

// GitHubCLISection captures GitHub CLI configuration and extensions.
type GitHubCLISection struct {
	ConfigDir  string   `toml:"config_dir,omitempty"`
	Extensions []string `toml:"extensions,omitempty"`
}

// GitReposSection captures discovered git repositories and their search paths.
type GitReposSection struct {
	SearchPaths  []string     `toml:"search_paths,omitempty"`
	Repositories []Repository `toml:"repositories,omitempty"`
}

// VSCodeSection captures VS Code extensions, config files, and snippets.
type VSCodeSection struct {
	Extensions  []string     `toml:"extensions,omitempty"`
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
	SnippetsDir string       `toml:"snippets_dir,omitempty"`
}

// CursorSection captures Cursor editor extensions and config files.
type CursorSection struct {
	Extensions  []string     `toml:"extensions,omitempty"`
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
}

// NeovimSection captures Neovim configuration directory and plugin manager.
type NeovimSection struct {
	ConfigDir     string `toml:"config_dir,omitempty"`
	PluginManager string `toml:"plugin_manager,omitempty"`
}

// JetBrainsSection captures JetBrains IDE installations and their settings.
type JetBrainsSection struct {
	IDEs []JetBrainsIDE `toml:"ides,omitempty"`
}

// XcodeSection captures Xcode simulators and config files.
type XcodeSection struct {
	Simulators  []string     `toml:"simulators,omitempty"`
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
}

// DockerSection captures Docker configuration, frequently used images, and runtime.
type DockerSection struct {
	ConfigFile           string   `toml:"config_file,omitempty"`
	FrequentlyUsedImages []string `toml:"frequently_used_images,omitempty"`
	Runtime              string   `toml:"runtime,omitempty"`
}

// AWSSection captures AWS CLI configuration and profiles.
type AWSSection struct {
	ConfigFile string   `toml:"config_file,omitempty"`
	Profiles   []string `toml:"profiles,omitempty"`
}

// KubernetesSection captures Kubernetes configuration and contexts.
type KubernetesSection struct {
	ConfigFile string   `toml:"config_file,omitempty"`
	Contexts   []string `toml:"contexts,omitempty"`
}

// TerraformSection captures Terraform CLI configuration.
type TerraformSection struct {
	ConfigFile string `toml:"config_file,omitempty"`
}

// VercelSection captures Vercel CLI configuration directory.
type VercelSection struct {
	ConfigDir string `toml:"config_dir,omitempty"`
}

// MacOSDefaultsSection captures macOS system preferences organized by category.
type MacOSDefaultsSection struct {
	Dock           *DockConfig           `toml:"dock,omitempty"`
	Finder         *FinderConfig         `toml:"finder,omitempty"`
	Keyboard       *KeyboardConfig       `toml:"keyboard,omitempty"`
	Trackpad       *TrackpadConfig       `toml:"trackpad,omitempty"`
	MissionControl *MissionControlConfig `toml:"mission_control,omitempty"`
	Spotlight      *SpotlightConfig      `toml:"spotlight,omitempty"`
	Screenshots    *ScreenshotsConfig    `toml:"screenshots,omitempty"`
	MenuBar        *MenuBarConfig        `toml:"menu_bar,omitempty"`
	Defaults       []MacDefault          `toml:"defaults,omitempty"`
}

// LocaleSection captures language, region, timezone, and computer name.
type LocaleSection struct {
	Language      string `toml:"language,omitempty"`
	Region        string `toml:"region,omitempty"`
	Timezone      string `toml:"timezone,omitempty"`
	ComputerName  string `toml:"computer_name,omitempty"`
	LocalHostname string `toml:"local_hostname,omitempty"`
}

// LoginItemsSection captures apps that start at login.
type LoginItemsSection struct {
	Apps []string `toml:"apps,omitempty"`
}

// HostsFileSection captures custom /etc/hosts entries.
type HostsFileSection struct {
	CustomEntries []HostEntry `toml:"custom_entries,omitempty"`
}

// AppsSection captures App Store installed applications.
type AppsSection struct {
	AppStore []InstalledApp `toml:"app_store,omitempty"`
}

// RaycastSection captures Raycast configuration export file.
type RaycastSection struct {
	ExportFile string `toml:"export_file,omitempty"`
}

// KarabinerSection captures Karabiner-Elements configuration directory.
type KarabinerSection struct {
	ConfigDir string `toml:"config_dir,omitempty"`
}

// RectangleSection captures Rectangle window manager configuration.
type RectangleSection struct {
	ConfigFile string `toml:"config_file,omitempty"`
}

// SSHSection captures SSH keys and configuration (typically encrypted).
type SSHSection struct {
	Encrypted  bool     `toml:"encrypted,omitempty"`
	ConfigFile string   `toml:"config_file,omitempty"`
	Keys       []string `toml:"keys,omitempty"`
	KnownHosts string   `toml:"known_hosts,omitempty"`
}

// GPGSection captures GPG keys and configuration (typically encrypted).
type GPGSection struct {
	Encrypted   bool         `toml:"encrypted,omitempty"`
	Keys        []string     `toml:"keys,omitempty"`
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
}

// XDGConfigSection captures auto-detected and custom XDG config paths.
type XDGConfigSection struct {
	AutoDetected []string `toml:"auto_detected,omitempty"`
	CustomPaths  []string `toml:"custom_paths,omitempty"`
	ConfigDir    string   `toml:"config_dir,omitempty"`
}

// FoldersSection captures workspace directory structure to be recreated.
type FoldersSection struct {
	Structure []string `toml:"structure,omitempty"`
}

// FontsSection captures user-installed custom fonts and Homebrew font casks.
type FontsSection struct {
	CustomFonts   []Font   `toml:"custom_fonts,omitempty"`
	HomebrewFonts []string `toml:"homebrew_fonts,omitempty"`
}

// EnvFilesSection captures environment files (typically encrypted).
type EnvFilesSection struct {
	Encrypted bool      `toml:"encrypted,omitempty"`
	Files     []EnvFile `toml:"files,omitempty"`
}

// CrontabSection captures the user's crontab entries.
type CrontabSection struct {
	Entries []string `toml:"entries,omitempty"`
}

// LaunchAgentsSection captures user-level LaunchAgent plists.
type LaunchAgentsSection struct {
	Plists []ConfigFile `toml:"plists,omitempty"`
}

// NetworkSection captures network preferences: Wi-Fi, DNS, and VPN configurations.
type NetworkSection struct {
	PreferredWifi []string     `toml:"preferred_wifi,omitempty"`
	DNS           *DNSConfig   `toml:"dns,omitempty"`
	VPNConfigs    []ConfigFile `toml:"vpn_configs,omitempty"`
}

// BrowserSection captures default browser and extensions checklist.
type BrowserSection struct {
	Default             string `toml:"default,omitempty"`
	ExtensionsChecklist string `toml:"extensions_checklist,omitempty"`
}

// AIToolsSection captures AI developer tool configurations.
type AIToolsSection struct {
	ClaudeCodeConfig string   `toml:"claude_code_config,omitempty"`
	OllamaModels     []string `toml:"ollama_models,omitempty"`
}

// APIToolsSection captures API and dev tool configurations.
type APIToolsSection struct {
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
	Mkcert      bool         `toml:"mkcert,omitempty"`
}

// DatabasesSection captures database client configuration files.
type DatabasesSection struct {
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
}

// RegistriesSection captures package registry configuration files.
type RegistriesSection struct {
	ConfigFiles []ConfigFile `toml:"config_files,omitempty"`
}

// DenoSection captures Deno version and globally installed packages.
type DenoSection struct {
	Version        string    `toml:"version,omitempty"`
	GlobalPackages []Package `toml:"global_packages,omitempty"`
}

// BunSection captures Bun version and globally installed packages.
type BunSection struct {
	Version        string    `toml:"version,omitempty"`
	GlobalPackages []Package `toml:"global_packages,omitempty"`
}

// RubySection captures Ruby version manager, versions, and global gems.
type RubySection struct {
	Manager        string    `toml:"manager,omitempty"`
	Versions       []string  `toml:"versions,omitempty"`
	DefaultVersion string    `toml:"default_version,omitempty"`
	GlobalGems     []Package `toml:"global_gems,omitempty"`
}

// GCPSection captures Google Cloud CLI configuration directory.
type GCPSection struct {
	ConfigDir string `toml:"config_dir,omitempty"`
}

// AzureSection captures Azure CLI configuration directory.
type AzureSection struct {
	ConfigDir string `toml:"config_dir,omitempty"`
}

// FlyioSection captures Fly.io CLI configuration file.
type FlyioSection struct {
	ConfigFile string `toml:"config_file,omitempty"`
}

// FirebaseSection captures Firebase CLI configuration directory.
type FirebaseSection struct {
	ConfigDir string `toml:"config_dir,omitempty"`
}

// CloudflareSection captures Cloudflare Wrangler configuration directory.
type CloudflareSection struct {
	ConfigDir string `toml:"config_dir,omitempty"`
}

// AlfredSection captures Alfred configuration directory.
type AlfredSection struct {
	ConfigDir string `toml:"config_dir,omitempty"`
}

// BetterTouchToolSection captures BetterTouchTool configuration file.
type BetterTouchToolSection struct {
	ConfigFile string `toml:"config_file,omitempty"`
}

// OnePasswordSection captures 1Password configuration directory.
type OnePasswordSection struct {
	ConfigDir string `toml:"config_dir,omitempty"`
}
