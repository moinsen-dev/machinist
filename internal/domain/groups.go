package domain

import "reflect"

// RestoreGroup describes a numbered group of restore stages.
type RestoreGroup struct {
	ID             string   // e.g. "01-foundation"
	Name           string   // e.g. "foundation"
	Label          string   // e.g. "Foundation"
	ScriptName     string   // e.g. "01-foundation.sh"
	SnapshotFields []string // Snapshot struct field names
}

// HasData returns true if the snapshot has any non-nil section for this group.
func (g RestoreGroup) HasData(snap *Snapshot) bool {
	v := reflect.ValueOf(snap).Elem()
	for _, fieldName := range g.SnapshotFields {
		f := v.FieldByName(fieldName)
		if f.IsValid() && f.Kind() == reflect.Ptr && !f.IsNil() {
			return true
		}
	}
	return false
}

// StageCount returns the number of non-nil stages in this group.
func (g RestoreGroup) StageCount(snap *Snapshot) int {
	count := 0
	v := reflect.ValueOf(snap).Elem()
	for _, fieldName := range g.SnapshotFields {
		f := v.FieldByName(fieldName)
		if f.IsValid() && f.Kind() == reflect.Ptr && !f.IsNil() {
			count++
		}
	}
	return count
}

var restoreGroups = []RestoreGroup{
	{ID: "01-homebrew", Name: "homebrew", Label: "Homebrew Packages", ScriptName: "01-homebrew.sh",
		SnapshotFields: []string{"Homebrew"}},
	{ID: "02-secrets", Name: "secrets", Label: "Secrets & Keys", ScriptName: "02-secrets.sh",
		SnapshotFields: []string{"SSH", "GPG", "EnvFiles"}},
	{ID: "03-configs", Name: "configs", Label: "Config Files", ScriptName: "03-configs.sh",
		SnapshotFields: []string{"Git", "GitHubCLI", "Shell", "Terminal", "Tmux",
			"VSCode", "Cursor", "Neovim", "JetBrains", "Xcode",
			"Docker", "AWS", "Kubernetes", "Terraform", "Vercel",
			"GCP", "Azure", "Flyio", "Firebase", "CloudflareWrangler",
			"Karabiner", "Rectangle", "BetterTouchTool", "OnePassword",
			"AITools", "APITools", "XDGConfig", "Databases", "Registries",
			"Fonts", "Folders", "Browser", "Raycast", "Alfred", "LoginItems"}},
	{ID: "04-runtimes", Name: "runtimes", Label: "Runtime Installers", ScriptName: "04-runtimes.sh",
		SnapshotFields: []string{"Node", "Python", "Rust", "Java", "Flutter", "Go",
			"Ruby", "Deno", "Bun", "Asdf"}},
	{ID: "05-repos", Name: "repos", Label: "Git Repositories", ScriptName: "05-repos.sh",
		SnapshotFields: []string{"GitRepos"}},
	{ID: "06-macos", Name: "macos", Label: "macOS System Settings", ScriptName: "06-macos.sh",
		SnapshotFields: []string{"MacOSDefaults", "Apps", "Crontab", "LaunchAgents",
			"Locale", "HostsFile", "Network"}},
}

// RestoreGroups returns all restore groups in order.
func RestoreGroups() []RestoreGroup { return restoreGroups }

// GroupByName finds a restore group by its short name.
func GroupByName(name string) (RestoreGroup, bool) {
	for _, g := range restoreGroups {
		if g.Name == name {
			return g, true
		}
	}
	return RestoreGroup{}, false
}

// GroupNames returns all group short names in order.
func GroupNames() []string {
	names := make([]string, len(restoreGroups))
	for i, g := range restoreGroups {
		names[i] = g.Name
	}
	return names
}
