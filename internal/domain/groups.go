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
	{ID: "01-foundation", Name: "foundation", Label: "Foundation", ScriptName: "01-foundation.sh",
		SnapshotFields: []string{"Homebrew", "SSH", "GPG", "Git"}},
	{ID: "02-shell", Name: "shell", Label: "Shell", ScriptName: "02-shell.sh",
		SnapshotFields: []string{"Shell", "Terminal", "Tmux"}},
	{ID: "03-runtimes", Name: "runtimes", Label: "Runtimes", ScriptName: "03-runtimes.sh",
		SnapshotFields: []string{"Node", "Python", "Rust", "Flutter", "Go", "Ruby", "Java", "Deno", "Bun", "Asdf"}},
	{ID: "04-editors", Name: "editors", Label: "Editors", ScriptName: "04-editors.sh",
		SnapshotFields: []string{"VSCode", "Cursor", "Xcode", "JetBrains", "Neovim", "GitHubCLI"}},
	{ID: "05-infrastructure", Name: "infrastructure", Label: "Infrastructure", ScriptName: "05-infrastructure.sh",
		SnapshotFields: []string{"Docker", "Kubernetes", "GCP", "AWS", "Azure", "Terraform", "Firebase", "CloudflareWrangler", "Vercel", "Flyio"}},
	{ID: "06-repos", Name: "repos", Label: "Repos", ScriptName: "06-repos.sh",
		SnapshotFields: []string{"GitRepos"}},
	{ID: "07-system", Name: "system", Label: "System", ScriptName: "07-system.sh",
		SnapshotFields: []string{"MacOSDefaults", "Fonts", "Apps", "Folders", "Crontab", "LaunchAgents",
			"Locale", "LoginItems", "HostsFile", "Network", "Raycast", "Alfred",
			"Karabiner", "Rectangle", "BetterTouchTool", "OnePassword", "Browser",
			"AITools", "APITools", "XDGConfig", "EnvFiles", "Databases", "Registries"}},
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
