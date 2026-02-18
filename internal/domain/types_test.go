package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSensitivity_String(t *testing.T) {
	tests := []struct {
		name     string
		s        Sensitivity
		expected string
	}{
		{"public level", Public, "public"},
		{"sensitive level", Sensitive, "sensitive"},
		{"secret level", Secret, "secret"},
		{"unknown level defaults to unknown", Sensitivity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.s.String())
		})
	}
}

func TestPackage_Fields(t *testing.T) {
	tests := []struct {
		name    string
		pkg     Package
		expName string
		expVer  string
	}{
		{
			name:    "package with name and version",
			pkg:     Package{Name: "git", Version: "2.43.0"},
			expName: "git",
			expVer:  "2.43.0",
		},
		{
			name:    "package with name only",
			pkg:     Package{Name: "ripgrep"},
			expName: "ripgrep",
			expVer:  "",
		},
		{
			name:    "empty package",
			pkg:     Package{},
			expName: "",
			expVer:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expName, tt.pkg.Name)
			assert.Equal(t, tt.expVer, tt.pkg.Version)
		})
	}
}

func TestServiceEntry_Fields(t *testing.T) {
	entry := ServiceEntry{Name: "postgresql@16", Status: "started"}
	assert.Equal(t, "postgresql@16", entry.Name)
	assert.Equal(t, "started", entry.Status)
}

func TestConfigFile_Fields(t *testing.T) {
	tests := []struct {
		name string
		cf   ConfigFile
	}{
		{
			name: "plain config file",
			cf: ConfigFile{
				Source:     "~/.zshrc",
				BundlePath: "configs/shell/.zshrc",
			},
		},
		{
			name: "encrypted config file",
			cf: ConfigFile{
				Source:     "~/.pgpass",
				BundlePath: "configs/db/.pgpass.age",
				Encrypted:  true,
			},
		},
		{
			name: "sensitive config file",
			cf: ConfigFile{
				Source:     "~/.npmrc",
				BundlePath: "configs/registries/.npmrc",
				Sensitive:  true,
			},
		},
		{
			name: "config file with content hash",
			cf: ConfigFile{
				Source:      "~/.gitconfig",
				BundlePath:  "configs/git/.gitconfig",
				ContentHash: "abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.cf.Source)
			assert.NotEmpty(t, tt.cf.BundlePath)
		})
	}
}

func TestRepository_Fields(t *testing.T) {
	repo := Repository{
		Path:    "~/Code/machinist",
		Remote:  "git@github.com:moinsen/machinist.git",
		Branch:  "main",
		Shallow: false,
	}
	assert.Equal(t, "~/Code/machinist", repo.Path)
	assert.Equal(t, "git@github.com:moinsen/machinist.git", repo.Remote)
	assert.Equal(t, "main", repo.Branch)
	assert.False(t, repo.Shallow)
}

func TestHostEntry_Fields(t *testing.T) {
	entry := HostEntry{
		IP:        "127.0.0.1",
		Hostnames: []string{"local.myapp.dev", "api.local.myapp.dev"},
	}
	assert.Equal(t, "127.0.0.1", entry.IP)
	assert.Len(t, entry.Hostnames, 2)
	assert.Contains(t, entry.Hostnames, "local.myapp.dev")
}

func TestMacDefault_Fields(t *testing.T) {
	d := MacDefault{
		Domain:    "com.apple.dock",
		Key:       "autohide",
		Value:     "true",
		ValueType: "bool",
	}
	assert.Equal(t, "com.apple.dock", d.Domain)
	assert.Equal(t, "autohide", d.Key)
	assert.Equal(t, "true", d.Value)
	assert.Equal(t, "bool", d.ValueType)
}

func TestInstalledApp_Fields(t *testing.T) {
	app := InstalledApp{
		Name:     "Xcode",
		Source:   "app_store",
		BundleID: "com.apple.dt.Xcode",
		ID:       497799835,
	}
	assert.Equal(t, "Xcode", app.Name)
	assert.Equal(t, "app_store", app.Source)
	assert.Equal(t, "com.apple.dt.Xcode", app.BundleID)
	assert.Equal(t, 497799835, app.ID)
}

func TestFont_Fields(t *testing.T) {
	f := Font{
		Name:       "FiraCode Nerd Font",
		BundlePath: "configs/fonts/FiraCode-NF.ttf",
	}
	assert.Equal(t, "FiraCode Nerd Font", f.Name)
	assert.Equal(t, "configs/fonts/FiraCode-NF.ttf", f.BundlePath)
}

func TestEnvFile_Fields(t *testing.T) {
	e := EnvFile{
		Source:     "~/Code/my-app/.env",
		BundlePath: "configs/env/my-app.env.age",
	}
	assert.Equal(t, "~/Code/my-app/.env", e.Source)
	assert.Equal(t, "configs/env/my-app.env.age", e.BundlePath)
}

func TestJetBrainsIDE_Fields(t *testing.T) {
	ide := JetBrainsIDE{
		Name:           "IntelliJ IDEA",
		SettingsExport: "configs/jetbrains/intellij-settings.zip",
	}
	assert.Equal(t, "IntelliJ IDEA", ide.Name)
	assert.Equal(t, "configs/jetbrains/intellij-settings.zip", ide.SettingsExport)
}

func TestAsdfPlugin_Fields(t *testing.T) {
	p := AsdfPlugin{
		Name:     "elixir",
		Versions: []string{"1.16.0"},
	}
	assert.Equal(t, "elixir", p.Name)
	assert.Len(t, p.Versions, 1)
}

func TestDNSConfig_Fields(t *testing.T) {
	dns := DNSConfig{
		Interface: "Wi-Fi",
		Servers:   []string{"1.1.1.1", "8.8.8.8"},
	}
	assert.Equal(t, "Wi-Fi", dns.Interface)
	assert.Len(t, dns.Servers, 2)
}

func TestDockConfig_Fields(t *testing.T) {
	dock := DockConfig{
		AutoHide:       true,
		TileSize:       48,
		Orientation:    "bottom",
		MinimizeEffect: "scale",
	}
	assert.True(t, dock.AutoHide)
	assert.Equal(t, 48, dock.TileSize)
	assert.Equal(t, "bottom", dock.Orientation)
}

func TestHotCorners_Fields(t *testing.T) {
	hc := HotCorners{
		TopLeft:     "mission-control",
		BottomRight: "desktop",
	}
	assert.Equal(t, "mission-control", hc.TopLeft)
	assert.Equal(t, "desktop", hc.BottomRight)
}
