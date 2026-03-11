# Split Restore Scripts Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use moinsenpowers:executing-plans to implement this plan task-by-task.

**Goal:** Split the monolithic `install.command` into 7 numbered group scripts that can be independently re-run, plus a thin orchestrator.

**Architecture:** Each group script is a standalone bash file with shared preamble (logging, run_stage, arch check). The orchestrator iterates over `0[1-7]-*.sh` and runs each. Group templates live in `templates/groups/` and compose existing stage templates. `GenerateRestoreScript()` becomes `GenerateRestoreScripts()` returning `map[string]string`.

**Tech Stack:** Go text/template, embed.FS, cobra CLI, bash shell scripts

---

## Group → Snapshot Field Mapping

| Group | Script | Snapshot Fields |
|-------|--------|----------------|
| foundation | `01-foundation.sh` | Homebrew, SSH, GPG, Git |
| shell | `02-shell.sh` | Shell, Terminal, Tmux |
| runtimes | `03-runtimes.sh` | Node, Python, Rust, Flutter, Go, Ruby, Java, Deno, Bun, Asdf |
| editors | `04-editors.sh` | VSCode, Cursor, Xcode, JetBrains, Neovim, GitHubCLI |
| infrastructure | `05-infrastructure.sh` | Docker, Kubernetes, GCP, AWS, Azure, Terraform, Firebase, CloudflareWrangler, Vercel, Flyio |
| repos | `06-repos.sh` | GitRepos |
| system | `07-system.sh` | MacOSDefaults, Fonts, Apps, Folders, Crontab, LaunchAgents, Locale, LoginItems, HostsFile, Network, Raycast, Alfred, Karabiner, Rectangle, BetterTouchTool, OnePassword, Browser, AITools, APITools, XDGConfig, EnvFiles, Databases, Registries |

---

### Task 1: Define group metadata in domain

**Files:**
- Create: `internal/domain/groups.go`
- Test: `internal/domain/groups_test.go`

**Step 1: Write the failing test**

```go
// internal/domain/groups_test.go
package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestoreGroups_OrderedByNumber(t *testing.T) {
	groups := RestoreGroups()
	require.Len(t, groups, 7)
	assert.Equal(t, "01-foundation", groups[0].ID)
	assert.Equal(t, "07-system", groups[6].ID)
}

func TestRestoreGroups_AllFieldsMapped(t *testing.T) {
	groups := RestoreGroups()
	// Every group must have at least one snapshot field
	for _, g := range groups {
		assert.NotEmpty(t, g.SnapshotFields, "group %s has no fields", g.ID)
	}
}

func TestGroupByName(t *testing.T) {
	g, ok := GroupByName("runtimes")
	require.True(t, ok)
	assert.Equal(t, "03-runtimes", g.ID)

	_, ok = GroupByName("nonexistent")
	assert.False(t, ok)
}

func TestGroupHasData_EmptySnapshot(t *testing.T) {
	snap := &Snapshot{Meta: Meta{}}
	groups := RestoreGroups()
	for _, g := range groups {
		assert.False(t, g.HasData(snap), "group %s should have no data on empty snapshot", g.ID)
	}
}

func TestGroupHasData_FoundationWithHomebrew(t *testing.T) {
	snap := &Snapshot{
		Meta:     Meta{},
		Homebrew: &HomebrewSection{},
	}
	g, _ := GroupByName("foundation")
	assert.True(t, g.HasData(snap))
}

func TestGroupHasData_RuntimesPartial(t *testing.T) {
	snap := &Snapshot{
		Meta: Meta{},
		Rust: &RustSection{},
	}
	g, _ := GroupByName("runtimes")
	assert.True(t, g.HasData(snap))
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/domain/ -run TestRestoreGroups -v`
Expected: FAIL — `RestoreGroups` not defined

**Step 3: Write minimal implementation**

```go
// internal/domain/groups.go
package domain

import "reflect"

// RestoreGroup describes a numbered group of restore stages.
type RestoreGroup struct {
	ID             string   // e.g. "01-foundation"
	Name           string   // e.g. "foundation"
	Label          string   // e.g. "Foundation"
	ScriptName     string   // e.g. "01-foundation.sh"
	SnapshotFields []string // Snapshot struct field names, e.g. ["Homebrew", "SSH", "GPG", "Git"]
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

// StageCount returns the number of non-nil stages in this group for the given snapshot.
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
	{
		ID: "01-foundation", Name: "foundation", Label: "Foundation",
		ScriptName:     "01-foundation.sh",
		SnapshotFields: []string{"Homebrew", "SSH", "GPG", "Git"},
	},
	{
		ID: "02-shell", Name: "shell", Label: "Shell",
		ScriptName:     "02-shell.sh",
		SnapshotFields: []string{"Shell", "Terminal", "Tmux"},
	},
	{
		ID: "03-runtimes", Name: "runtimes", Label: "Runtimes",
		ScriptName:     "03-runtimes.sh",
		SnapshotFields: []string{"Node", "Python", "Rust", "Flutter", "Go", "Ruby", "Java", "Deno", "Bun", "Asdf"},
	},
	{
		ID: "04-editors", Name: "editors", Label: "Editors",
		ScriptName:     "04-editors.sh",
		SnapshotFields: []string{"VSCode", "Cursor", "Xcode", "JetBrains", "Neovim", "GitHubCLI"},
	},
	{
		ID: "05-infrastructure", Name: "infrastructure", Label: "Infrastructure",
		ScriptName:     "05-infrastructure.sh",
		SnapshotFields: []string{"Docker", "Kubernetes", "GCP", "AWS", "Azure", "Terraform", "Firebase", "CloudflareWrangler", "Vercel", "Flyio"},
	},
	{
		ID: "06-repos", Name: "repos", Label: "Repos",
		ScriptName:     "06-repos.sh",
		SnapshotFields: []string{"GitRepos"},
	},
	{
		ID: "07-system", Name: "system", Label: "System",
		ScriptName: "07-system.sh",
		SnapshotFields: []string{
			"MacOSDefaults", "Fonts", "Apps", "Folders", "Crontab", "LaunchAgents",
			"Locale", "LoginItems", "HostsFile", "Network", "Raycast", "Alfred",
			"Karabiner", "Rectangle", "BetterTouchTool", "OnePassword", "Browser",
			"AITools", "APITools", "XDGConfig", "EnvFiles", "Databases", "Registries",
		},
	},
}

// RestoreGroups returns the ordered list of restore groups.
func RestoreGroups() []RestoreGroup {
	return restoreGroups
}

// GroupByName looks up a group by its short name (e.g. "foundation").
func GroupByName(name string) (RestoreGroup, bool) {
	for _, g := range restoreGroups {
		if g.Name == name {
			return g, true
		}
	}
	return RestoreGroup{}, false
}

// GroupNames returns the short names of all groups, in order.
func GroupNames() []string {
	names := make([]string, len(restoreGroups))
	for i, g := range restoreGroups {
		names[i] = g.Name
	}
	return names
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/domain/ -run "TestRestoreGroups|TestGroupByName|TestGroupHasData" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/groups.go internal/domain/groups_test.go
git commit -m "feat: add RestoreGroup metadata with snapshot field mapping"
```

---

### Task 2: Create shared preamble template

**Files:**
- Create: `templates/groups/_preamble.sh.tmpl`

**Step 1: Create the preamble template**

This is a named template fragment that each group script will include. It receives two template arguments via a wrapper struct: the snapshot and group-specific metadata.

```bash
{{/* templates/groups/_preamble.sh.tmpl */}}
{{define "preamble"}}#!/bin/bash
set -uo pipefail

cd "$(dirname "$0")"

# machinist restore: {{.GroupLabel}}
# Generated: {{.Meta.CreatedAt}}
# Source: {{.Meta.SourceHostname}} ({{.Meta.SourceOSVersion}}, {{.Meta.SourceArch}})

LOGFILE="$HOME/.machinist/restore-{{.GroupID}}.log"
mkdir -p "$(dirname "$LOGFILE")"
STAGE_NUM=0
STAGE_TOTAL={{.StageCount}}
STAGE_PASS=0
STAGE_FAIL=0
STAGE_SKIP=0

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOGFILE"; }

stage() {
    STAGE_NUM=$((STAGE_NUM + 1))
    log "[$STAGE_NUM/$STAGE_TOTAL] === $1 ==="
}

run_stage() {
    local name="$1"
    shift
    stage "$name"
    if "$@"; then
        STAGE_PASS=$((STAGE_PASS + 1))
        log "  -> $name completed"
    else
        STAGE_FAIL=$((STAGE_FAIL + 1))
        log "  !! $name failed (continuing)"
    fi
}

CURRENT_ARCH=$(uname -m)
SOURCE_ARCH="{{.Meta.SourceArch}}"
if [ "$CURRENT_ARCH" != "$SOURCE_ARCH" ] && [ -n "$SOURCE_ARCH" ]; then
    log "WARNING: Architecture mismatch: snapshot from $SOURCE_ARCH, running on $CURRENT_ARCH"
    log "  Some packages may need Rosetta 2 or different versions"
fi

START_TIME=$(date +%s)
log "{{.GroupLabel}} restore started"
{{end}}
```

**Step 2: No test needed — this is a template fragment tested through Task 4**

**Step 3: Commit**

```bash
git add templates/groups/_preamble.sh.tmpl
git commit -m "feat: add shared preamble template for group scripts"
```

---

### Task 3: Create 7 group templates

**Files:**
- Create: `templates/groups/01-foundation.sh.tmpl`
- Create: `templates/groups/02-shell.sh.tmpl`
- Create: `templates/groups/03-runtimes.sh.tmpl`
- Create: `templates/groups/04-editors.sh.tmpl`
- Create: `templates/groups/05-infrastructure.sh.tmpl`
- Create: `templates/groups/06-repos.sh.tmpl`
- Create: `templates/groups/07-system.sh.tmpl`

Each group template follows the same pattern — preamble + conditional stages + summary.

**Step 1: Create `templates/groups/01-foundation.sh.tmpl`**

```
{{template "preamble" .}}

{{if .Homebrew}}
do_homebrew() {
{{template "homebrew" .Homebrew}}
}
run_stage "Homebrew" do_homebrew
{{end}}

{{if .SSH}}
do_ssh() {
{{template "ssh" .SSH}}
}
run_stage "SSH Keys" do_ssh
{{end}}

{{if .GPG}}
do_gpg() {
{{template "gpg" .GPG}}
}
run_stage "GPG Keys" do_gpg
{{end}}

{{if .Git}}
do_git_config() {
{{template "git-config" .Git}}
}
run_stage "Git Configuration" do_git_config
{{end}}

{{template "summary" .}}
```

**Step 2: Create `templates/groups/02-shell.sh.tmpl`**

```
{{template "preamble" .}}

{{if .Shell}}
do_shell() {
{{template "shell" .Shell}}
}
run_stage "Shell Configuration" do_shell
{{end}}

{{if .Terminal}}
do_terminal() {
{{template "terminal" .Terminal}}
}
run_stage "Terminal Emulator" do_terminal
{{end}}

{{if .Tmux}}
do_tmux() {
{{template "tmux" .Tmux}}
}
run_stage "tmux" do_tmux
{{end}}

{{template "summary" .}}
```

**Step 3: Create `templates/groups/03-runtimes.sh.tmpl`**

```
{{template "preamble" .}}

{{if .Node}}
do_node() {
{{template "node" .Node}}
}
run_stage "Node.js" do_node
{{end}}

{{if .Python}}
do_python() {
{{template "python" .Python}}
}
run_stage "Python" do_python
{{end}}

{{if .Rust}}
do_rust() {
{{template "rust" .Rust}}
}
run_stage "Rust" do_rust
{{end}}

{{if .Flutter}}
do_flutter() {
{{template "flutter" .Flutter}}
}
run_stage "Flutter" do_flutter
{{end}}

{{if .Go}}
do_go() {
{{template "go-runtime" .Go}}
}
run_stage "Go" do_go
{{end}}

{{if .Ruby}}
do_ruby() {
{{template "ruby" .Ruby}}
}
run_stage "Ruby" do_ruby
{{end}}

{{if .Java}}
do_java() {
{{template "java" .Java}}
}
run_stage "Java/SDKMAN" do_java
{{end}}

{{if .Deno}}
do_deno() {
{{template "deno" .Deno}}
}
run_stage "Deno" do_deno
{{end}}

{{if .Bun}}
do_bun() {
{{template "bun" .Bun}}
}
run_stage "Bun" do_bun
{{end}}

{{if .Asdf}}
do_asdf() {
{{template "asdf" .Asdf}}
}
run_stage "asdf/mise" do_asdf
{{end}}

{{template "summary" .}}
```

**Step 4: Create `templates/groups/04-editors.sh.tmpl`**

```
{{template "preamble" .}}

{{if .VSCode}}
do_vscode() {
{{template "vscode" .VSCode}}
}
run_stage "Visual Studio Code" do_vscode
{{end}}

{{if .Cursor}}
do_cursor() {
{{template "cursor" .Cursor}}
}
run_stage "Cursor" do_cursor
{{end}}

{{if .Xcode}}
do_xcode() {
{{template "xcode" .Xcode}}
}
run_stage "Xcode" do_xcode
{{end}}

{{if .JetBrains}}
do_jetbrains() {
{{template "jetbrains" .JetBrains}}
}
run_stage "JetBrains IDEs" do_jetbrains
{{end}}

{{if .Neovim}}
do_neovim() {
{{template "neovim" .Neovim}}
}
run_stage "Neovim" do_neovim
{{end}}

{{if .GitHubCLI}}
do_github_cli() {
{{template "github-cli" .GitHubCLI}}
}
run_stage "GitHub CLI" do_github_cli
{{end}}

{{template "summary" .}}
```

**Step 5: Create `templates/groups/05-infrastructure.sh.tmpl`**

```
{{template "preamble" .}}

{{if .Docker}}
do_docker() {
{{template "docker" .Docker}}
}
run_stage "Docker" do_docker
{{end}}

{{if .Kubernetes}}
do_kubernetes() {
{{template "kubernetes" .Kubernetes}}
}
run_stage "Kubernetes" do_kubernetes
{{end}}

{{if .GCP}}
do_gcp() {
{{template "gcp" .GCP}}
}
run_stage "Google Cloud" do_gcp
{{end}}

{{if .AWS}}
do_aws() {
{{template "aws" .AWS}}
}
run_stage "AWS CLI" do_aws
{{end}}

{{if .Azure}}
do_azure() {
{{template "azure" .Azure}}
}
run_stage "Azure" do_azure
{{end}}

{{if .Terraform}}
do_terraform() {
{{template "terraform" .Terraform}}
}
run_stage "Terraform" do_terraform
{{end}}

{{if .Firebase}}
do_firebase() {
{{template "firebase" .Firebase}}
}
run_stage "Firebase" do_firebase
{{end}}

{{if .CloudflareWrangler}}
do_cloudflare() {
{{template "cloudflare" .CloudflareWrangler}}
}
run_stage "Cloudflare" do_cloudflare
{{end}}

{{if .Vercel}}
do_vercel() {
{{template "vercel" .Vercel}}
}
run_stage "Vercel" do_vercel
{{end}}

{{if .Flyio}}
do_flyio() {
{{template "flyio" .Flyio}}
}
run_stage "Fly.io" do_flyio
{{end}}

{{template "summary" .}}
```

**Step 6: Create `templates/groups/06-repos.sh.tmpl`**

```
{{template "preamble" .}}

{{if .GitRepos}}
do_git_repos() {
{{template "git-repos" .GitRepos}}
}
run_stage "Git Repositories" do_git_repos
{{end}}

{{template "summary" .}}
```

**Step 7: Create `templates/groups/07-system.sh.tmpl`**

```
{{template "preamble" .}}

{{if .MacOSDefaults}}
do_macos_defaults() {
{{template "macos-defaults" .MacOSDefaults}}
}
run_stage "macOS Defaults" do_macos_defaults
{{end}}

{{if .Apps}}
do_apps() {
{{template "apps" .Apps}}
}
run_stage "Mac App Store Apps" do_apps
{{end}}

{{if .Fonts}}
do_fonts() {
{{template "fonts" .Fonts}}
}
run_stage "Fonts" do_fonts
{{end}}

{{if .Folders}}
do_folders() {
{{template "folders" .Folders}}
}
run_stage "Folder Structure" do_folders
{{end}}

{{if or .Crontab .LaunchAgents}}
do_scheduled() {
{{template "scheduled" .}}
}
run_stage "Scheduled Tasks" do_scheduled
{{end}}

{{if .Locale}}
do_locale() {
{{template "locale" .Locale}}
}
run_stage "Locale & Timezone" do_locale
{{end}}

{{if .LoginItems}}
do_login_items() {
{{template "login-items" .LoginItems}}
}
run_stage "Login Items" do_login_items
{{end}}

{{if .HostsFile}}
do_hosts_file() {
{{template "hosts-file" .HostsFile}}
}
run_stage "Hosts File" do_hosts_file
{{end}}

{{if .Network}}
do_network() {
{{template "network" .Network}}
}
run_stage "Network" do_network
{{end}}

{{if .Raycast}}
do_raycast() {
{{template "raycast" .Raycast}}
}
run_stage "Raycast" do_raycast
{{end}}

{{if .Alfred}}
do_alfred() {
{{template "alfred" .Alfred}}
}
run_stage "Alfred" do_alfred
{{end}}

{{if .Karabiner}}
do_karabiner() {
{{template "karabiner" .Karabiner}}
}
run_stage "Karabiner-Elements" do_karabiner
{{end}}

{{if .Rectangle}}
do_rectangle() {
{{template "rectangle" .Rectangle}}
}
run_stage "Rectangle" do_rectangle
{{end}}

{{if .BetterTouchTool}}
do_bettertouchtool() {
{{template "bettertouchtool" .BetterTouchTool}}
}
run_stage "BetterTouchTool" do_bettertouchtool
{{end}}

{{if .OnePassword}}
do_onepassword() {
{{template "onepassword" .OnePassword}}
}
run_stage "1Password CLI" do_onepassword
{{end}}

{{if .Browser}}
do_browser() {
{{template "browser" .Browser}}
}
run_stage "Browser" do_browser
{{end}}

{{if .AITools}}
do_ai_tools() {
{{template "ai-tools" .AITools}}
}
run_stage "AI Tools" do_ai_tools
{{end}}

{{if .APITools}}
do_api_tools() {
{{template "api-tools" .APITools}}
}
run_stage "API Tools" do_api_tools
{{end}}

{{if .XDGConfig}}
do_xdg_config() {
{{template "xdg-config" .XDGConfig}}
}
run_stage "XDG Config" do_xdg_config
{{end}}

{{if .EnvFiles}}
do_env_files() {
{{template "env-files" .EnvFiles}}
}
run_stage "Environment Files" do_env_files
{{end}}

{{if .Databases}}
do_databases() {
{{template "databases" .Databases}}
}
run_stage "Database Clients" do_databases
{{end}}

{{if .Registries}}
do_registries() {
{{template "registries" .Registries}}
}
run_stage "Package Registries" do_registries
{{end}}

{{template "summary" .}}
```

**Step 8: Create shared summary template fragment**

Add to `templates/groups/_preamble.sh.tmpl` (append):

```
{{define "summary"}}
END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
log ""
log "{{.GroupLabel}} restore completed in ${ELAPSED}s"
log "  Passed: $STAGE_PASS stages succeeded"
log "  Failed: $STAGE_FAIL stages failed"
log "  Skipped: $STAGE_SKIP stages skipped"
echo ""
echo "Check $LOGFILE for details."
if [ $STAGE_FAIL -gt 0 ]; then exit 1; fi
{{end}}
```

**Step 9: Commit**

```bash
git add templates/groups/
git commit -m "feat: add 7 group templates with shared preamble and summary"
```

---

### Task 4: Update embed directive and add GroupTemplateData

**Files:**
- Modify: `templates.go`
- Create: `internal/bundler/groupdata.go`
- Test: `internal/bundler/groupdata_test.go`

**Step 1: Update `templates.go` to include group templates**

```go
// templates.go
package machinist

import "embed"

//go:embed templates/*.tmpl templates/stages/*.tmpl templates/groups/*.tmpl
var TemplateFS embed.FS
```

**Step 2: Write the failing test for GroupTemplateData**

The group templates use `.GroupLabel`, `.GroupID`, `.StageCount`, `.Meta`, and all snapshot fields. We need a wrapper struct.

```go
// internal/bundler/groupdata_test.go
package bundler

import (
	"testing"
	"time"

	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewGroupTemplateData(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: domain.Meta{
			CreatedAt:      time.Now(),
			SourceHostname: "test-mac",
			SourceArch:     "arm64",
		},
		Homebrew: &domain.HomebrewSection{},
		SSH:      &domain.SSHSection{},
	}
	group, _ := domain.GroupByName("foundation")

	data := NewGroupTemplateData(snap, group)

	assert.Equal(t, "Foundation", data.GroupLabel)
	assert.Equal(t, "01-foundation", data.GroupID)
	assert.Equal(t, 2, data.StageCount) // Homebrew + SSH
	assert.NotNil(t, data.Homebrew)
	assert.NotNil(t, data.SSH)
	assert.Equal(t, "test-mac", data.Meta.SourceHostname)
}
```

**Step 3: Run test to verify it fails**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -run TestNewGroupTemplateData -v`
Expected: FAIL — `NewGroupTemplateData` not defined

**Step 4: Write the implementation**

```go
// internal/bundler/groupdata.go
package bundler

import "github.com/moinsen-dev/machinist/internal/domain"

// GroupTemplateData embeds the Snapshot and adds group-specific fields
// needed by group templates (GroupLabel, GroupID, StageCount).
type GroupTemplateData struct {
	*domain.Snapshot
	GroupLabel string
	GroupID    string
	StageCount int
}

// NewGroupTemplateData creates the template data for a group.
func NewGroupTemplateData(snap *domain.Snapshot, group domain.RestoreGroup) GroupTemplateData {
	return GroupTemplateData{
		Snapshot:   snap,
		GroupLabel: group.Label,
		GroupID:    group.ID,
		StageCount: group.StageCount(snap),
	}
}
```

**Step 5: Run test to verify it passes**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -run TestNewGroupTemplateData -v`
Expected: PASS

**Step 6: Commit**

```bash
git add templates.go internal/bundler/groupdata.go internal/bundler/groupdata_test.go
git commit -m "feat: add GroupTemplateData and update embed to include group templates"
```

---

### Task 5: Implement GenerateRestoreScripts

**Files:**
- Modify: `internal/bundler/restorescript.go`
- Modify: `internal/bundler/restorescript_test.go`

**Step 1: Write the failing test**

```go
// Add to internal/bundler/restorescript_test.go

func TestGenerateRestoreScripts_ReturnsMapOfScripts(t *testing.T) {
	snap := &domain.Snapshot{
		Meta:     newMeta(),
		Homebrew: &domain.HomebrewSection{Formulae: []domain.Package{{Name: "git"}}},
		Shell:    &domain.ShellSection{DefaultShell: "/bin/zsh"},
		Node:     &domain.NodeSection{Version: "20.0.0"},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	// Should have orchestrator + 3 group scripts (foundation, shell, runtimes)
	assert.Contains(t, scripts, "install.command")
	assert.Contains(t, scripts, "01-foundation.sh")
	assert.Contains(t, scripts, "02-shell.sh")
	assert.Contains(t, scripts, "03-runtimes.sh")

	// Groups without data should NOT be present
	assert.NotContains(t, scripts, "04-editors.sh")
	assert.NotContains(t, scripts, "05-infrastructure.sh")
	assert.NotContains(t, scripts, "06-repos.sh")
	assert.NotContains(t, scripts, "07-system.sh")
}

func TestGenerateRestoreScripts_GroupScriptIsStandalone(t *testing.T) {
	snap := &domain.Snapshot{
		Meta:     newMeta(),
		Homebrew: &domain.HomebrewSection{Formulae: []domain.Package{{Name: "git"}}},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	foundation := scripts["01-foundation.sh"]
	assert.Contains(t, foundation, "#!/bin/bash")
	assert.Contains(t, foundation, "set -uo pipefail")
	assert.Contains(t, foundation, `LOGFILE="$HOME/.machinist/restore-01-foundation.log"`)
	assert.Contains(t, foundation, `run_stage "Homebrew"`)
	assert.Contains(t, foundation, "brew install git")
	assert.Contains(t, foundation, "Foundation restore completed")
}

func TestGenerateRestoreScripts_OrchestratorRunsGroupsInOrder(t *testing.T) {
	snap := &domain.Snapshot{
		Meta:     newMeta(),
		Homebrew: &domain.HomebrewSection{},
		Shell:    &domain.ShellSection{},
		GitRepos: &domain.GitReposSection{},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	orchestrator := scripts["install.command"]
	assert.Contains(t, orchestrator, "#!/bin/bash")
	assert.Contains(t, orchestrator, "01-foundation.sh")
	assert.Contains(t, orchestrator, "02-shell.sh")
	assert.Contains(t, orchestrator, "06-repos.sh")

	// Verify order
	idx01 := strings.Index(orchestrator, "01-foundation.sh")
	idx02 := strings.Index(orchestrator, "02-shell.sh")
	idx06 := strings.Index(orchestrator, "06-repos.sh")
	assert.Less(t, idx01, idx02)
	assert.Less(t, idx02, idx06)
}

func TestGenerateRestoreScripts_EmptySnapshot(t *testing.T) {
	snap := &domain.Snapshot{Meta: newMeta()}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	// Only orchestrator, no group scripts
	assert.Len(t, scripts, 1)
	assert.Contains(t, scripts, "install.command")
}

func TestGenerateRestoreScripts_GroupStageCount(t *testing.T) {
	snap := &domain.Snapshot{
		Meta:     newMeta(),
		Homebrew: &domain.HomebrewSection{},
		SSH:      &domain.SSHSection{Keys: []string{"id_ed25519"}},
		GPG:      &domain.GPGSection{},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	foundation := scripts["01-foundation.sh"]
	// 3 stages: Homebrew, SSH, GPG
	assert.Contains(t, foundation, "STAGE_TOTAL=3")
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -run TestGenerateRestoreScripts -v`
Expected: FAIL — `GenerateRestoreScripts` not defined

**Step 3: Write the implementation**

Add to `internal/bundler/restorescript.go`:

```go
// GenerateRestoreScripts renders the split restore scripts from a Snapshot.
// Returns a map of filename → content. Always includes "install.command" (orchestrator).
// Group scripts (e.g. "01-foundation.sh") are only included when the snapshot
// has data for that group.
func GenerateRestoreScripts(snapshot *domain.Snapshot) (map[string]string, error) {
	funcMap := template.FuncMap{
		"base": filepath.Base,
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(
		machinist.TemplateFS,
		"templates/*.tmpl",
		"templates/stages/*.tmpl",
		"templates/groups/*.tmpl",
	)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	scripts := make(map[string]string)
	var groupScriptNames []string

	for _, group := range domain.RestoreGroups() {
		if !group.HasData(snapshot) {
			continue
		}

		data := NewGroupTemplateData(snapshot, group)
		var buf bytes.Buffer
		templateName := group.ScriptName + ".tmpl"
		if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
			return nil, fmt.Errorf("execute template %s: %w", templateName, err)
		}
		scripts[group.ScriptName] = buf.String()
		groupScriptNames = append(groupScriptNames, group.ScriptName)
	}

	// Generate orchestrator
	orchestrator := generateOrchestrator(snapshot, groupScriptNames)
	scripts["install.command"] = orchestrator

	return scripts, nil
}

// generateOrchestrator creates the thin install.command that runs group scripts sequentially.
func generateOrchestrator(snapshot *domain.Snapshot, scriptNames []string) string {
	var buf bytes.Buffer
	buf.WriteString("#!/bin/bash\n")
	buf.WriteString("set -uo pipefail\n\n")
	buf.WriteString("cd \"$(dirname \"$0\")\"\n\n")
	buf.WriteString(fmt.Sprintf("# machinist restore orchestrator\n"))
	buf.WriteString(fmt.Sprintf("# Generated: %s\n", snapshot.Meta.CreatedAt.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("# Source: %s (%s, %s)\n\n", snapshot.Meta.SourceHostname, snapshot.Meta.SourceOSVersion, snapshot.Meta.SourceArch))
	buf.WriteString("TOTAL=0\nPASSED=0\nFAILED=0\n\n")
	buf.WriteString("echo \"machinist restore — running group scripts\"\necho \"\"\n\n")

	for _, name := range scriptNames {
		buf.WriteString(fmt.Sprintf("if [ -f \"%s\" ]; then\n", name))
		buf.WriteString(fmt.Sprintf("    echo \">>> Running %s\"\n", name))
		buf.WriteString(fmt.Sprintf("    TOTAL=$((TOTAL + 1))\n"))
		buf.WriteString(fmt.Sprintf("    if bash \"%s\"; then\n", name))
		buf.WriteString(fmt.Sprintf("        PASSED=$((PASSED + 1))\n"))
		buf.WriteString(fmt.Sprintf("    else\n"))
		buf.WriteString(fmt.Sprintf("        FAILED=$((FAILED + 1))\n"))
		buf.WriteString(fmt.Sprintf("        echo \"!!! %s failed (continuing)\"\n", name))
		buf.WriteString(fmt.Sprintf("    fi\n"))
		buf.WriteString(fmt.Sprintf("    echo \"\"\n"))
		buf.WriteString(fmt.Sprintf("fi\n\n"))
	}

	buf.WriteString("echo \"machinist restore complete: $PASSED/$TOTAL groups passed")
	buf.WriteString(", $FAILED failed\"\n")
	buf.WriteString("if [ $FAILED -gt 0 ]; then exit 1; fi\n")

	return buf.String()
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -run TestGenerateRestoreScripts -v`
Expected: PASS

**Step 5: Verify existing tests still pass**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -v`
Expected: All PASS (old `GenerateRestoreScript` tests still work)

**Step 6: Commit**

```bash
git add internal/bundler/restorescript.go internal/bundler/restorescript_test.go
git commit -m "feat: add GenerateRestoreScripts returning split group scripts"
```

---

### Task 6: Update PrepareBundleDir to write multiple scripts

**Files:**
- Modify: `internal/bundler/dmg.go:26-47`

**Step 1: Write the failing test**

```go
// Add to internal/bundler/dmg_test.go (or create if not exists)

func TestPrepareBundleDir_WritesGroupScripts(t *testing.T) {
	snap := &domain.Snapshot{
		Meta:     newTestMeta(),
		Homebrew: &domain.HomebrewSection{Formulae: []domain.Package{{Name: "git"}}},
		Shell:    &domain.ShellSection{DefaultShell: "/bin/zsh"},
	}

	outputDir := t.TempDir()
	err := PrepareBundleDir(snap, outputDir, t.TempDir(), "")
	require.NoError(t, err)

	// Orchestrator
	assertFileExists(t, filepath.Join(outputDir, "install.command"))

	// Group scripts
	assertFileExists(t, filepath.Join(outputDir, "01-foundation.sh"))
	assertFileExists(t, filepath.Join(outputDir, "02-shell.sh"))

	// Groups without data should NOT be written
	assertFileNotExists(t, filepath.Join(outputDir, "03-runtimes.sh"))
	assertFileNotExists(t, filepath.Join(outputDir, "04-editors.sh"))

	// Group scripts should be executable
	info, _ := os.Stat(filepath.Join(outputDir, "01-foundation.sh"))
	assert.True(t, info.Mode()&0111 != 0, "group script should be executable")
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, "expected file to exist: %s", path)
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), "expected file to NOT exist: %s", path)
}

func newTestMeta() domain.Meta {
	return domain.Meta{
		CreatedAt:      time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		SourceHostname: "test-mac",
		SourceOSVersion: "darwin",
		SourceArch:     "arm64",
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -run TestPrepareBundleDir_WritesGroupScripts -v`
Expected: FAIL — still writes single install.command

**Step 3: Update PrepareBundleDir**

In `internal/bundler/dmg.go`, replace the block at lines 39-47 (the single install.command generation) with:

```go
	// Generate and write split restore scripts
	scripts, err := GenerateRestoreScripts(snapshot)
	if err != nil {
		return fmt.Errorf("generate restore scripts: %w", err)
	}
	for filename, content := range scripts {
		perm := os.FileMode(0755) // all scripts are executable
		if filepath.Ext(filename) == ".md" {
			perm = 0644
		}
		path := filepath.Join(outputDir, filename)
		if err := os.WriteFile(path, []byte(content), perm); err != nil {
			return fmt.Errorf("write %s: %w", filename, err)
		}
	}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -run TestPrepareBundleDir -v`
Expected: PASS

**Step 5: Run all bundler tests**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/bundler/dmg.go internal/bundler/dmg_test.go
git commit -m "feat: PrepareBundleDir writes split group scripts instead of monolithic install.command"
```

---

### Task 7: Update restore CLI for group-based --only/--skip

**Files:**
- Modify: `cmd/machinist/restore.go`
- Modify: `cmd/machinist/restore_test.go`

**Step 1: Write the failing tests**

```go
// Add to cmd/machinist/restore_test.go

func TestRestore_ListGroups(t *testing.T) {
	cmd := rootCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"restore", "--list"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "foundation")
	assert.Contains(t, output, "runtimes")
	assert.Contains(t, output, "system")
}

func TestRestore_OnlyAcceptsGroupNames(t *testing.T) {
	// Create a temp manifest
	tmpDir := t.TempDir()
	snap := &domain.Snapshot{
		Meta:     domain.Meta{SourceHostname: "test"},
		Homebrew: &domain.HomebrewSection{},
		Shell:    &domain.ShellSection{},
	}
	domain.WriteManifest(snap, filepath.Join(tmpDir, "manifest.toml"))

	cmd := rootCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"restore", filepath.Join(tmpDir, "manifest.toml"), "--only", "foundation", "--dry-run"})
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "foundation")
	assert.NotContains(t, output, "shell")
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./cmd/machinist/ -run "TestRestore_ListGroups|TestRestore_OnlyAcceptsGroupNames" -v`
Expected: FAIL

**Step 3: Rewrite restore.go for group-based execution**

```go
// cmd/machinist/restore.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/moinsen-dev/machinist/internal/bundler"
	"github.com/moinsen-dev/machinist/internal/domain"
	"github.com/spf13/cobra"
)

var (
	restoreSkip    string
	restoreOnly    string
	restoreDryRun  bool
	restoreYes     bool
	restoreList    bool
)

var restoreCmd = &cobra.Command{
	Use:   "restore [manifest.toml]",
	Short: "Restore environment from manifest",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if restoreList {
			fmt.Fprintln(cmd.OutOrStdout(), "Available restore groups:")
			for _, g := range domain.RestoreGroups() {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", g.Name, g.Label)
			}
			return nil
		}

		manifestPath := "manifest.toml"
		if len(args) > 0 {
			manifestPath = args[0]
		}

		if restoreSkip != "" && restoreOnly != "" {
			return fmt.Errorf("--skip and --only are mutually exclusive; use one or the other")
		}

		snap, err := domain.ReadManifest(manifestPath)
		if err != nil {
			return fmt.Errorf("read manifest: %w", err)
		}

		// Determine which groups to run
		groups := domain.RestoreGroups()
		var selectedGroups []domain.RestoreGroup

		if restoreOnly != "" {
			only := parseCSV(restoreOnly)
			onlySet := make(map[string]bool, len(only))
			for _, n := range only {
				onlySet[n] = true
			}
			for _, g := range groups {
				if onlySet[g.Name] && g.HasData(snap) {
					selectedGroups = append(selectedGroups, g)
				}
			}
		} else if restoreSkip != "" {
			skip := parseCSV(restoreSkip)
			skipSet := make(map[string]bool, len(skip))
			for _, n := range skip {
				skipSet[n] = true
			}
			for _, g := range groups {
				if !skipSet[g.Name] && g.HasData(snap) {
					selectedGroups = append(selectedGroups, g)
				}
			}
		} else {
			for _, g := range groups {
				if g.HasData(snap) {
					selectedGroups = append(selectedGroups, g)
				}
			}
		}

		if restoreDryRun {
			fmt.Fprintf(cmd.OutOrStdout(), "Dry-run mode: restore plan\n")
			fmt.Fprintf(cmd.OutOrStdout(), "Manifest: %s\n", manifestPath)
			fmt.Fprintf(cmd.OutOrStdout(), "Host: %s (%s)\n", snap.Meta.SourceHostname, snap.Meta.SourceArch)
			fmt.Fprintf(cmd.OutOrStdout(), "Groups to execute: %d\n", len(selectedGroups))
			for i, g := range selectedGroups {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s (%d stages)\n", i+1, g.Name, g.StageCount(snap))
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nNo changes were made (dry-run).")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Restoring %d groups from %s\n", len(selectedGroups), manifestPath)
		for _, g := range selectedGroups {
			fmt.Fprintf(cmd.OutOrStdout(), "  - %s (%d stages)\n", g.Name, g.StageCount(snap))
		}

		if !restoreYes {
			fmt.Fprintln(cmd.OutOrStdout(), "\nUse --yes to confirm execution.")
			return nil
		}

		// Execute group scripts
		bundleDir := filepath.Dir(manifestPath)
		for _, g := range selectedGroups {
			scriptPath := filepath.Join(bundleDir, g.ScriptName)
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				// Generate script on-the-fly if not found in bundle
				scripts, genErr := bundler.GenerateRestoreScripts(snap)
				if genErr != nil {
					return fmt.Errorf("generate restore scripts: %w", genErr)
				}
				content, ok := scripts[g.ScriptName]
				if !ok {
					continue
				}
				scriptPath = filepath.Join(os.TempDir(), g.ScriptName)
				if writeErr := os.WriteFile(scriptPath, []byte(content), 0755); writeErr != nil {
					return fmt.Errorf("write temp script: %w", writeErr)
				}
				defer os.Remove(scriptPath)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "\n>>> Running %s\n", g.ScriptName)
			execCmd := exec.CommandContext(cmd.Context(), "bash", scriptPath)
			execCmd.Dir = bundleDir
			execCmd.Stdout = cmd.OutOrStdout()
			execCmd.Stderr = cmd.ErrOrStderr()
			if err := execCmd.Run(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "!!! %s failed: %v (continuing)\n", g.ScriptName, err)
			}
		}

		fmt.Fprintln(cmd.OutOrStdout(), "\nRestore complete.")
		return nil
	},
}

func init() {
	restoreCmd.Flags().StringVar(&restoreSkip, "skip", "", "Comma-separated group names to skip (foundation,shell,runtimes,editors,infrastructure,repos,system)")
	restoreCmd.Flags().StringVar(&restoreOnly, "only", "", "Comma-separated group names to run (exclusive with --skip)")
	restoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "Show what would be executed without doing it")
	restoreCmd.Flags().BoolVarP(&restoreYes, "yes", "y", false, "Skip confirmation prompt")
	restoreCmd.Flags().BoolVar(&restoreList, "list", false, "List available restore groups")
	rootCmd.AddCommand(restoreCmd)
}
```

**Step 4: Run tests**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./cmd/machinist/ -run TestRestore -v`
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/machinist/restore.go cmd/machinist/restore_test.go
git commit -m "feat: restore CLI uses group-based --only/--skip/--list"
```

---

### Task 8: Update existing restore tests for backward compatibility

**Files:**
- Modify: `cmd/machinist/restore_test.go`

**Step 1: Review and update existing tests**

The existing tests use `sectionNames()` for `--only`/`--skip`. Since we now use group names instead of section names, update the tests:

- `TestRestore_SkipAndOnlyMutuallyExclusive` — unchanged (still tests flag conflict)
- `TestRestore_NonExistentManifest` — unchanged
- `TestRestore_DryRun` — update to check group output format
- `TestRestore_WithoutYesFlag` — update to check group output format

**Step 2: Run all tests**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./... -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add cmd/machinist/restore_test.go
git commit -m "fix: update restore tests for group-based execution"
```

---

### Task 9: Handle `scheduled` template special case

**Files:**
- Modify: `templates/groups/07-system.sh.tmpl` (already created in Task 3)

The `scheduled` stage template uses `{{template "scheduled" .}}` where `.` is the full Snapshot (it accesses both `.Crontab` and `.LaunchAgents`). In the group template, `.` is `GroupTemplateData` which embeds `*Snapshot`, so `{{template "scheduled" .Snapshot}}` or simply `.` (since embed promotes fields) should work. But the `{{if or .Crontab .LaunchAgents}}` guard accesses promoted fields, so it works as-is.

**Step 1: Write a test to verify**

```go
// Add to internal/bundler/restorescript_test.go

func TestGenerateRestoreScripts_ScheduledTasksInSystem(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Crontab: &domain.CrontabSection{
			Entries: []domain.CrontabEntry{{Schedule: "0 * * * *", Command: "backup.sh"}},
		},
	}

	scripts, err := GenerateRestoreScripts(snap)
	require.NoError(t, err)

	system := scripts["07-system.sh"]
	require.NotEmpty(t, system)
	assert.Contains(t, system, `run_stage "Scheduled Tasks"`)
}
```

**Step 2: Run test**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./internal/bundler/ -run TestGenerateRestoreScripts_ScheduledTasks -v`
Expected: PASS (or FAIL if the `scheduled` template needs `.Snapshot` — fix accordingly)

**Step 3: If the `scheduled` template fails**

The `scheduled` template receives the full data context. Since `GroupTemplateData` embeds `*Snapshot`, the `.Crontab` and `.LaunchAgents` fields are promoted. However, if the template uses `{{.}}` as a receiver that must be `*Snapshot`, we'll need to pass `.Snapshot` instead:

Change in `07-system.sh.tmpl`:
```
{{if or .Crontab .LaunchAgents}}
do_scheduled() {
{{template "scheduled" .Snapshot}}
}
run_stage "Scheduled Tasks" do_scheduled
{{end}}
```

**Step 4: Commit if changes needed**

```bash
git add templates/groups/07-system.sh.tmpl internal/bundler/restorescript_test.go
git commit -m "fix: scheduled template receives Snapshot for Crontab+LaunchAgents access"
```

---

### Task 10: Clean up — remove sectionNames dependency from restore

**Files:**
- Modify: `cmd/machinist/snapshot.go` (sectionNames stays — used by snapshot command)
- Verify: `cmd/machinist/restore.go` no longer imports `reflect`

**Step 1: Verify restore.go doesn't use sectionNames**

The rewritten restore.go (Task 7) uses `domain.RestoreGroups()` instead of `sectionNames()`. Verify `sectionNames` is only used by `snapshot.go`.

**Step 2: Run full test suite**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go test ./... -count=1`
Expected: All PASS

**Step 3: Commit if any cleanup needed**

```bash
git add -A
git commit -m "chore: clean up unused imports after group-based restore migration"
```

---

### Task 11: End-to-end verification

**Step 1: Build**

Run: `cd /Users/udi/work/moinsen/opensource/machinist && go build ./cmd/machinist/`

**Step 2: Create a test snapshot and DMG**

Run: `./machinist snapshot -o /tmp/test-snap.toml`

Run: `./machinist dmg /tmp/test-snap.toml -o /tmp/test.dmg`

**Step 3: Mount and verify DMG contents**

Run: `hdiutil attach /tmp/test.dmg -mountpoint /tmp/test-mount`

Verify files exist:
```bash
ls -la /tmp/test-mount/machinist/
# Expected: install.command, 01-foundation.sh through 07-system.sh (only those with data),
#           manifest.toml, README.md, POST_RESTORE_CHECKLIST.md, configs/
```

**Step 4: Test standalone group execution**

Run: `bash /tmp/test-mount/machinist/01-foundation.sh` (or dry-read it)

**Step 5: Test restore CLI with groups**

Run: `./machinist restore --list`
Run: `./machinist restore /tmp/test-mount/machinist/manifest.toml --only foundation --dry-run`
Run: `./machinist restore /tmp/test-mount/machinist/manifest.toml --skip repos,system --dry-run`

**Step 6: Clean up**

Run: `hdiutil detach /tmp/test-mount && rm /tmp/test.dmg /tmp/test-snap.toml`

**Step 7: Final commit**

Only commit if any fixes were needed during verification.
