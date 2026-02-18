package domain

// Package represents a software package with an optional version.
type Package struct {
	Name    string `toml:"name"`
	Version string `toml:"version,omitempty"`
}

// ServiceEntry represents a background service managed by a package manager.
type ServiceEntry struct {
	Name   string `toml:"name"`
	Status string `toml:"status"`
}

// ConfigFile represents a configuration file to be captured and bundled.
type ConfigFile struct {
	Source      string `toml:"source"`
	BundlePath  string `toml:"bundle_path"`
	ContentHash string `toml:"content_hash,omitempty"`
	Encrypted   bool   `toml:"encrypted,omitempty"`
	Sensitive   bool   `toml:"sensitive,omitempty"`
}

// Repository represents a git repository to be cloned during restore.
type Repository struct {
	Path    string `toml:"path"`
	Remote  string `toml:"remote"`
	Branch  string `toml:"branch"`
	Shallow bool   `toml:"shallow,omitempty"`
}

// HostEntry represents a custom /etc/hosts entry.
type HostEntry struct {
	IP        string   `toml:"ip"`
	Hostnames []string `toml:"hostnames"`
}

// MacDefault represents a macOS defaults write entry.
type MacDefault struct {
	Domain    string `toml:"domain"`
	Key       string `toml:"key"`
	Value     string `toml:"value"`
	ValueType string `toml:"value_type"`
}

// InstalledApp represents an application installed from the App Store or other sources.
type InstalledApp struct {
	Name     string `toml:"name"`
	Source   string `toml:"source,omitempty"`
	BundleID string `toml:"bundle_id,omitempty"`
	ID       int    `toml:"id,omitempty"`
}

// Font represents a user-installed font.
type Font struct {
	Name       string `toml:"name"`
	BundlePath string `toml:"bundle_path"`
}

// EnvFile represents an environment file to be captured (typically encrypted).
type EnvFile struct {
	Source     string `toml:"source"`
	BundlePath string `toml:"bundle_path"`
}

// JetBrainsIDE represents a JetBrains IDE installation with its settings export.
type JetBrainsIDE struct {
	Name           string `toml:"name"`
	SettingsExport string `toml:"settings_export,omitempty"`
}

// AsdfPlugin represents an asdf/mise plugin with its installed versions.
type AsdfPlugin struct {
	Name     string   `toml:"name"`
	Versions []string `toml:"versions"`
}

// DNSConfig represents DNS server configuration for a network interface.
type DNSConfig struct {
	Interface string   `toml:"interface"`
	Servers   []string `toml:"servers"`
}

// DockConfig represents macOS Dock preferences.
type DockConfig struct {
	AutoHide       bool   `toml:"autohide,omitempty"`
	TileSize       int    `toml:"tilesize,omitempty"`
	Orientation    string `toml:"orientation,omitempty"`
	MinimizeEffect string `toml:"minimize_effect,omitempty"`
}

// FinderConfig represents macOS Finder preferences.
type FinderConfig struct {
	ShowHidden    bool   `toml:"show_hidden,omitempty"`
	ShowPathBar   bool   `toml:"show_path_bar,omitempty"`
	ShowStatusBar bool   `toml:"show_status_bar,omitempty"`
	DefaultView   string `toml:"default_view,omitempty"`
}

// KeyboardConfig represents macOS keyboard preferences.
type KeyboardConfig struct {
	KeyRepeat        int `toml:"key_repeat,omitempty"`
	InitialKeyRepeat int `toml:"initial_key_repeat,omitempty"`
}

// TrackpadConfig represents macOS trackpad preferences.
type TrackpadConfig struct {
	TapToClick    bool    `toml:"tap_to_click,omitempty"`
	TrackingSpeed float64 `toml:"tracking_speed,omitempty"`
}

// MissionControlConfig represents macOS Mission Control preferences.
type MissionControlConfig struct {
	HotCorners *HotCorners `toml:"hot_corners,omitempty"`
}

// HotCorners represents macOS hot corner assignments.
type HotCorners struct {
	TopLeft     string `toml:"top_left,omitempty"`
	TopRight    string `toml:"top_right,omitempty"`
	BottomLeft  string `toml:"bottom_left,omitempty"`
	BottomRight string `toml:"bottom_right,omitempty"`
}

// SpotlightConfig represents macOS Spotlight preferences.
type SpotlightConfig struct {
	ExcludedPaths []string `toml:"excluded_paths,omitempty"`
}

// ScreenshotsConfig represents macOS screenshot preferences.
type ScreenshotsConfig struct {
	Path           string `toml:"path,omitempty"`
	Format         string `toml:"format,omitempty"`
	DisableShadow  bool   `toml:"disable_shadow,omitempty"`
}

// MenuBarConfig represents macOS menu bar preferences.
type MenuBarConfig struct {
	ClockFormat           string `toml:"clock_format,omitempty"`
	ShowBatteryPercentage bool   `toml:"show_battery_percentage,omitempty"`
}
