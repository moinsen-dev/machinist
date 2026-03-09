# Restore Template Fixes — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use moinsenpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix all template-vs-bundler path mismatches and stage ordering issues so the DMG restore script is fully self-installable on a vanilla Mac.

**Architecture:** Templates must use hardcoded bundle-relative paths (e.g. `configs/docker/config.json`) instead of raw domain field values (e.g. `{{.ConfigFile}}`). The Homebrew stage needs PATH initialization for Apple Silicon. Stage ordering must put SSH/GPG before Git repos. Fonts need bundling support.

**Tech Stack:** Go, text/template, TDD with testify

---

### Task 1: Fix Homebrew PATH initialization for Apple Silicon

The Homebrew installer puts `brew` at `/opt/homebrew/bin/brew` on arm64 Macs, which is not on the default `$PATH`. Every subsequent `brew install` call fails silently.

**Files:**
- Modify: `templates/stages/homebrew.sh.tmpl`
- Modify: `internal/bundler/restorescript_test.go`

**Step 1: Write the failing test**

Add to `restorescript_test.go`:

```go
func TestGenerateRestoreScript_HomebrewPATHInit(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		Homebrew: &domain.HomebrewSection{
			Formulae: []domain.Package{{Name: "git"}},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// After installing Homebrew, the script must initialize the PATH
	// so that brew is available in the current shell session.
	assert.Contains(t, script, `/opt/homebrew/bin/brew shellenv`)
	assert.Contains(t, script, `/usr/local/bin/brew shellenv`)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_HomebrewPATHInit -v`
Expected: FAIL — the template doesn't contain shellenv lines yet.

**Step 3: Fix the template**

In `templates/stages/homebrew.sh.tmpl`, add PATH initialization right after the Homebrew install block (after the `fi` on line 6):

```bash
{{define "homebrew"}}
# Install Homebrew if not present
if ! command -v brew &>/dev/null; then
    log "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

# Ensure brew is on PATH for the rest of this script (critical on Apple Silicon)
eval "$(/opt/homebrew/bin/brew shellenv)" 2>/dev/null || eval "$(/usr/local/bin/brew shellenv)" 2>/dev/null || true

{{range .Taps}}
log "Tapping {{.}}"
brew tap {{.}} 2>/dev/null || true
{{end}}

{{range .Formulae}}
log "Installing formula {{.Name}}"
brew list {{.Name}} &>/dev/null || brew install {{.Name}}
{{end}}

{{range .Casks}}
log "Installing cask {{.Name}}"
brew list --cask {{.Name}} &>/dev/null || brew install --cask {{.Name}}
{{end}}

{{range .Services}}
{{if eq .Status "started"}}
log "Starting service {{.Name}}"
brew services start {{.Name}} 2>/dev/null || true
{{end}}
{{end}}
{{end}}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_HomebrewPATHInit -v`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./internal/bundler/ -v`
Expected: All existing tests still pass.

**Step 6: Commit**

```bash
git add templates/stages/homebrew.sh.tmpl internal/bundler/restorescript_test.go
git commit -m "fix: add Homebrew PATH initialization for Apple Silicon"
```

---

### Task 2: Fix stage ordering — SSH and GPG before Git Repos

SSH keys must be restored before `git clone` (stage 3 currently), since private repo clones need SSH. GPG should also come before Git config since signing may be configured.

**Files:**
- Modify: `templates/install.command.tmpl`
- Modify: `internal/bundler/restorescript_test.go`

**Step 1: Write the failing test**

```go
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

	// SSH must appear before Git Repos
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_SSHBeforeGitRepos -v`
Expected: FAIL — SSH currently comes after Git Repos.

**Step 3: Reorder stages in install.command.tmpl**

Move the stage blocks to this order (after Homebrew):

1. Homebrew (stays first)
2. SSH Keys (moved up from position ~15)
3. GPG Keys (moved up from position ~16)
4. Git Configuration (moved up from position ~14)
5. Shell Configuration
6. Git Repositories
7. ...everything else stays in current relative order...

In `install.command.tmpl`, the new order after the Homebrew block should be:

```
{{if .SSH}} ... SSH Keys ... {{end}}
{{if .GPG}} ... GPG Keys ... {{end}}
{{if .Git}} ... Git Configuration ... {{end}}
{{if .Shell}} ... Shell Configuration ... {{end}}
{{if .GitRepos}} ... Git Repositories ... {{end}}
{{if .Node}} ... Node.js ... {{end}}
...rest unchanged...
```

Remove the SSH, GPG, and Git Configuration blocks from their old positions (lines 142-161).

**Step 4: Run test to verify it passes**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_SSHBeforeGitRepos -v`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./internal/bundler/ -v`
Expected: All existing tests still pass.

**Step 6: Commit**

```bash
git add templates/install.command.tmpl internal/bundler/restorescript_test.go
git commit -m "fix: reorder stages — SSH/GPG/Git config before Git repos"
```

---

### Task 3: Fix GPG bundle path mismatch (`configs/gnupg/` → `configs/gpg/`)

The GPG template expects `configs/gnupg/` but the bundler writes to `configs/gpg/`. Fix the template to match the bundler.

**Files:**
- Modify: `templates/stages/security.sh.tmpl`
- Modify: `internal/bundler/restorescript_test.go`

**Step 1: Write the failing test**

```go
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

	// Template must use configs/gpg/ (matching bundler), not configs/gnupg/
	assert.Contains(t, script, `configs/gpg/ABC123`)
	assert.NotContains(t, script, `configs/gnupg/`)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_GPGBundlePath -v`
Expected: FAIL — template currently has `configs/gnupg/`.

**Step 3: Fix the template**

In `templates/stages/security.sh.tmpl`, replace all occurrences of `configs/gnupg/` with `configs/gpg/`:

Line 66: `if [ -f "configs/gpg/{{.}}.asc.age" ]; then`
Line 68: `age --decrypt "configs/gpg/{{.}}.asc.age" <<< "$AGE_PASSPHRASE" | gpg --import || true`
Line 74: `if [ -f "configs/gpg/{{.}}.asc" ]; then`
Line 76: `gpg --import "configs/gpg/{{.}}.asc" || true`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_GPGBundlePath -v`
Expected: PASS

**Step 5: Commit**

```bash
git add templates/stages/security.sh.tmpl internal/bundler/restorescript_test.go
git commit -m "fix: GPG template path configs/gnupg → configs/gpg to match bundler"
```

---

### Task 4: Fix SSH config/known_hosts template paths

The SSH template uses `{{.ConfigFile}}` and `{{.KnownHosts}}` raw (source paths like `.ssh/config`), but the bundler puts them at `configs/ssh/config` and `configs/ssh/known_hosts`.

**Files:**
- Modify: `templates/stages/security.sh.tmpl`
- Modify: `internal/bundler/restorescript_test.go`

**Step 1: Write the failing test**

```go
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

	// Template must reference bundle paths, not source paths
	assert.Contains(t, script, `if [ -f "configs/ssh/config" ]`)
	assert.Contains(t, script, `cp "configs/ssh/config" "$HOME/.ssh/config"`)
	assert.Contains(t, script, `if [ -f "configs/ssh/known_hosts" ]`)
	assert.Contains(t, script, `cp "configs/ssh/known_hosts" "$HOME/.ssh/known_hosts"`)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_SSHConfigBundlePaths -v`
Expected: FAIL

**Step 3: Fix the template**

In `templates/stages/security.sh.tmpl`, replace the SSH config/known_hosts blocks (lines 32-46):

```
{{if .ConfigFile}}
if [ -f "configs/ssh/config" ]; then
    log "Restoring SSH config"
    cp "configs/ssh/config" "$HOME/.ssh/config"
    chmod 644 "$HOME/.ssh/config"
fi
{{end}}

{{if .KnownHosts}}
if [ -f "configs/ssh/known_hosts" ]; then
    log "Restoring known_hosts"
    cp "configs/ssh/known_hosts" "$HOME/.ssh/known_hosts"
    chmod 644 "$HOME/.ssh/known_hosts"
fi
{{end}}
```

**Step 4: Run tests**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_SSHConfig -v`
Expected: PASS

**Step 5: Commit**

```bash
git add templates/stages/security.sh.tmpl internal/bundler/restorescript_test.go
git commit -m "fix: SSH template uses bundle paths instead of raw source paths"
```

---

### Task 5: Fix single-string ConfigFile templates (Docker, AWS, K8s, Terraform, Flyio, Rectangle, BTT)

All these templates use `{{.ConfigFile}}` which is the source path (e.g. `.docker/config.json`). The bundler wraps them to `configs/<prefix>/<basename>`. Fix each template to use the known bundle path.

**Files:**
- Modify: `templates/stages/cloud.sh.tmpl`
- Modify: `templates/stages/tools.sh.tmpl`
- Modify: `internal/bundler/restorescript_test.go`

**Step 1: Write the failing test**

```go
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

	// Each template must use configs/<prefix>/<basename>, not the raw source path
	assert.Contains(t, script, `"configs/docker/config.json"`)
	assert.Contains(t, script, `"configs/aws/config"`)
	assert.Contains(t, script, `"configs/kubernetes/config"`)
	assert.Contains(t, script, `"configs/terraform/.terraformrc"`)
	assert.Contains(t, script, `"configs/flyio/config.yml"`)
	assert.Contains(t, script, `"configs/rectangle/com.knollsoft.Rectangle.plist"`)
	assert.Contains(t, script, `"configs/bettertouchtool/btt_data.json"`)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_ConfigFileBundlePaths -v`
Expected: FAIL

**Step 3: Fix each template**

**cloud.sh.tmpl — Docker:**
```
{{define "docker"}}
log "Restoring Docker..."
{{if .ConfigFile}}
if [ -f "configs/docker/$(basename "{{.ConfigFile}}")" ]; then
    log "Restoring Docker config"
    mkdir -p "$HOME/.docker"
    cp "configs/docker/$(basename "{{.ConfigFile}}")" "$HOME/.docker/config.json"
fi
{{end}}
```

Wait — this is fragile. The bundler always uses `filepath.Base(source)` inside the `wrap()` helper. The cleanest approach: use a Go template function to compute `configs/<prefix>/basename`. But that would require modifying `restorescript.go` to add custom template functions.

**Simpler approach:** Since the bundler's `wrap()` function deterministically creates `configs/<prefix>/<basename>`, and we know the prefix for each section, hardcode the bundle path pattern in each template. The `{{.ConfigFile}}` field tells us the *source* path, so we still need it for the destination (`$HOME/...`). But for the *bundle check/copy source*, we use the known pattern.

Actually, the cleanest fix: **add a template function `bundlePath`** that computes `configs/<prefix>/<basename>` from the source path. But even simpler: since each template is section-specific and knows its own prefix, we can just hardcode:

**cloud.sh.tmpl changes:**

Docker template (line 3-8):
```
{{if .ConfigFile}}
if [ -f "configs/docker/{{base .ConfigFile}}" ]; then
    log "Restoring Docker config"
    mkdir -p "$HOME/.docker"
    cp "configs/docker/{{base .ConfigFile}}" "$HOME/.docker/config.json"
fi
{{end}}
```

Hmm, but `{{base .ConfigFile}}` isn't a built-in Go template function. We'd need to register it. Let's do that.

**Better approach — register `base` as a template function in `restorescript.go`:**

In `restorescript.go`, change `template.ParseFS` to:

```go
funcMap := template.FuncMap{
    "base": filepath.Base,
}
tmpl, err := template.New("").Funcs(funcMap).ParseFS(machinist.TemplateFS, "templates/*.tmpl", "templates/stages/*.tmpl")
```

Then in each template, use `{{base .ConfigFile}}` to get just the filename.

**cloud.sh.tmpl — all sections updated:**

Docker:
```
{{define "docker"}}
log "Restoring Docker..."
{{if .ConfigFile}}
if [ -f "configs/docker/{{base .ConfigFile}}" ]; then
    log "Restoring Docker config"
    mkdir -p "$HOME/.docker"
    cp "configs/docker/{{base .ConfigFile}}" "$HOME/.docker/config.json"
fi
{{end}}
...
{{end}}
```

AWS:
```
{{define "aws"}}
log "Restoring AWS CLI..."
{{if .ConfigFile}}
if [ -f "configs/aws/{{base .ConfigFile}}" ]; then
    log "Restoring AWS config"
    mkdir -p "$HOME/.aws"
    cp "configs/aws/{{base .ConfigFile}}" "$HOME/.aws/config"
    chmod 600 "$HOME/.aws/config"
fi
{{end}}
...
{{end}}
```

Kubernetes:
```
{{define "kubernetes"}}
log "Restoring Kubernetes..."
{{if .ConfigFile}}
if [ -f "configs/kubernetes/{{base .ConfigFile}}" ]; then
    log "Restoring kubeconfig"
    mkdir -p "$HOME/.kube"
    cp "configs/kubernetes/{{base .ConfigFile}}" "$HOME/.kube/config"
    chmod 600 "$HOME/.kube/config"
fi
{{end}}
...
{{end}}
```

Terraform:
```
{{define "terraform"}}
log "Restoring Terraform..."
{{if .ConfigFile}}
if [ -f "configs/terraform/{{base .ConfigFile}}" ]; then
    log "Restoring Terraform CLI config"
    cp "configs/terraform/{{base .ConfigFile}}" "$HOME/.terraformrc"
fi
{{end}}
{{end}}
```

Flyio:
```
{{define "flyio"}}
log "Restoring Fly.io..."
{{if .ConfigFile}}
if [ -f "configs/flyio/{{base .ConfigFile}}" ]; then
    log "Restoring Fly.io config"
    mkdir -p "$HOME/.fly"
    cp "configs/flyio/{{base .ConfigFile}}" "$HOME/.fly/config.yml"
fi
{{end}}
log "CHECKLIST: Run 'fly auth login' to re-authenticate"
{{end}}
```

**tools.sh.tmpl — Rectangle:**
```
{{define "rectangle"}}
log "Restoring Rectangle..."
{{if .ConfigFile}}
if [ -f "configs/rectangle/{{base .ConfigFile}}" ]; then
    log "Restoring Rectangle preferences"
    cp "configs/rectangle/{{base .ConfigFile}}" "$HOME/Library/Preferences/com.knollsoft.Rectangle.plist"
    defaults read com.knollsoft.Rectangle &>/dev/null || true
fi
{{end}}
{{end}}
```

**tools.sh.tmpl — BetterTouchTool:**
```
{{define "bettertouchtool"}}
log "Restoring BetterTouchTool..."
{{if .ConfigFile}}
log "CHECKLIST: BetterTouchTool configuration"
log "  1. Open BetterTouchTool"
log "  2. Go to Preferences > Manage Presets"
log "  3. Import from: configs/bettertouchtool/{{base .ConfigFile}}"
{{end}}
{{end}}
```

**tools.sh.tmpl — Raycast (uses ExportFile):**
```
{{define "raycast"}}
log "Restoring Raycast..."
{{if .ExportFile}}
log "CHECKLIST: Raycast configuration export available"
log "  1. Open Raycast"
log "  2. Go to Settings > Advanced > Import"
log "  3. Import from: configs/raycast/{{base .ExportFile}}"
{{end}}
{{end}}
```

**tools.sh.tmpl — AI Tools:**
```
{{define "ai-tools"}}
log "Restoring AI tools..."
{{if .ClaudeCodeConfig}}
if [ -f "configs/ai-tools/{{base .ClaudeCodeConfig}}" ]; then
    log "Restoring Claude Code config"
    mkdir -p "$HOME/.claude"
    cp "configs/ai-tools/{{base .ClaudeCodeConfig}}" "$HOME/.claude/{{base .ClaudeCodeConfig}}"
fi
{{end}}
...
{{end}}
```

**Step 4: Register `base` function in restorescript.go**

Change line 15-18 in `internal/bundler/restorescript.go`:

```go
func GenerateRestoreScript(snapshot *domain.Snapshot) (string, error) {
	funcMap := template.FuncMap{
		"base": filepath.Base,
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(machinist.TemplateFS, "templates/*.tmpl", "templates/stages/*.tmpl")
	if err != nil {
		return "", fmt.Errorf("parse templates: %w", err)
	}
	...
}
```

Add `"path/filepath"` to imports.

**Step 5: Run tests**

Run: `go test ./internal/bundler/ -v`
Expected: All pass including the new test.

**Step 6: Commit**

```bash
git add internal/bundler/restorescript.go templates/stages/cloud.sh.tmpl templates/stages/tools.sh.tmpl internal/bundler/restorescript_test.go
git commit -m "fix: ConfigFile templates use bundle paths via base() function"
```

---

### Task 6: Fix ConfigDir templates (GitHub CLI, Neovim, Vercel, Firebase, Cloudflare, Karabiner, Alfred, 1Password)

These templates use `{{.ConfigDir}}` (source path like `.config/gh`), but the bundler writes to `configs/<prefix>/`. Fix each to use the known bundle path.

**Files:**
- Modify: `templates/stages/git.sh.tmpl`
- Modify: `templates/stages/cloud.sh.tmpl`
- Modify: `templates/stages/tools.sh.tmpl`
- Modify: `internal/bundler/restorescript_test.go`

**Step 1: Write the failing test**

```go
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

	// Each template must use configs/<prefix>/ bundle paths
	assert.Contains(t, script, `"configs/github-cli/"`)
	assert.Contains(t, script, `"configs/neovim/"`)
	assert.Contains(t, script, `"configs/vercel/"`)
	assert.Contains(t, script, `"configs/firebase/"`)
	assert.Contains(t, script, `"configs/cloudflare/"`)
	assert.Contains(t, script, `"configs/karabiner/"`)
	assert.Contains(t, script, `"configs/alfred/"`)
	assert.Contains(t, script, `"configs/onepassword/"`)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_ConfigDirBundlePaths -v`
Expected: FAIL

**Step 3: Fix each template**

**git.sh.tmpl — GitHub CLI:** Replace `{{.ConfigDir}}` with `configs/github-cli`:
```
{{define "github-cli"}}
log "Restoring GitHub CLI..."
if ! command -v gh &>/dev/null; then
    log "Warning: GitHub CLI is not installed — install via 'brew install gh'"
else
    {{if .ConfigDir}}
    if [ -d "configs/github-cli" ]; then
        log "Restoring GitHub CLI config directory"
        mkdir -p "$HOME/.config/gh"
        cp -R "configs/github-cli/" "$HOME/.config/gh/"
    fi
    {{end}}

    {{range .Extensions}}
    log "Installing gh extension {{.}}"
    gh extension install "{{.}}" || true
    {{end}}

    log "CHECKLIST: Run 'gh auth login' to authenticate"
fi
{{end}}
```

**cloud.sh.tmpl — Vercel:**
```
{{define "vercel"}}
log "Restoring Vercel..."
{{if .ConfigDir}}
if [ -d "configs/vercel" ]; then
    log "Restoring Vercel config directory"
    mkdir -p "$HOME/.config/com.vercel.cli"
    cp -R "configs/vercel/" "$HOME/.config/com.vercel.cli/"
fi
{{end}}
log "CHECKLIST: Run 'vercel login' to authenticate"
{{end}}
```

**cloud.sh.tmpl — Firebase:**
```
{{define "firebase"}}
log "Restoring Firebase..."
{{if .ConfigDir}}
if [ -d "configs/firebase" ]; then
    log "Restoring Firebase config directory"
    mkdir -p "$HOME/.config/firebase"
    cp -R "configs/firebase/" "$HOME/.config/firebase/"
fi
{{end}}
log "CHECKLIST: Run 'firebase login' to re-authenticate"
{{end}}
```

**cloud.sh.tmpl — Cloudflare:**
```
{{define "cloudflare"}}
log "Restoring Cloudflare..."
{{if .ConfigDir}}
if [ -d "configs/cloudflare" ]; then
    log "Restoring Cloudflare/Wrangler config directory"
    mkdir -p "$HOME/.config/.wrangler"
    cp -R "configs/cloudflare/" "$HOME/.config/.wrangler/"
fi
{{end}}
log "CHECKLIST: Run 'wrangler login' to re-authenticate"
{{end}}
```

**tools.sh.tmpl — Karabiner:**
```
{{define "karabiner"}}
log "Restoring Karabiner-Elements..."
{{if .ConfigDir}}
if [ -d "configs/karabiner" ]; then
    log "Restoring Karabiner config directory"
    mkdir -p "$HOME/.config/karabiner"
    cp -R "configs/karabiner/" "$HOME/.config/karabiner/"
fi
{{end}}
{{end}}
```

**tools.sh.tmpl — Alfred:**
```
{{define "alfred"}}
log "Restoring Alfred..."
{{if .ConfigDir}}
log "CHECKLIST: Alfred preferences sync"
log "  1. Open Alfred Preferences"
log "  2. Go to Advanced > Set preferences folder"
log "  3. Point to your synced preferences location (Dropbox/iCloud)"
log "  Config bundled at: configs/alfred/"
{{end}}
{{end}}
```

**tools.sh.tmpl — 1Password:**
```
{{define "onepassword"}}
log "Restoring 1Password CLI..."
# Install 1Password CLI if not present
if ! command -v op &>/dev/null; then
    log "Installing 1Password CLI..."
    brew install --cask 1password-cli || true
fi

{{if .ConfigDir}}
if [ -d "configs/onepassword" ]; then
    log "Restoring 1Password CLI config directory"
    mkdir -p "$HOME/.config/op"
    cp -R "configs/onepassword/" "$HOME/.config/op/"
fi
{{end}}
log "CHECKLIST: Run 'op signin' to authenticate with 1Password"
{{end}}
```

**editors_extra.sh.tmpl — Neovim** (need to check this file):

Read `templates/stages/editors_extra.sh.tmpl` to find the neovim template, then fix it:
```
{{define "neovim"}}
log "Restoring Neovim..."
{{if .ConfigDir}}
if [ -d "configs/neovim" ]; then
    log "Restoring Neovim config directory"
    mkdir -p "$HOME/.config/nvim"
    cp -R "configs/neovim/" "$HOME/.config/nvim/"
fi
{{end}}
{{if .PluginManager}}
log "Neovim plugin manager: {{.PluginManager}}"
log "CHECKLIST: Open Neovim and let plugins install"
{{end}}
{{end}}
```

**Step 4: Run tests**

Run: `go test ./internal/bundler/ -v`
Expected: All pass.

**Step 5: Commit**

```bash
git add templates/stages/git.sh.tmpl templates/stages/cloud.sh.tmpl templates/stages/tools.sh.tmpl templates/stages/editors_extra.sh.tmpl internal/bundler/restorescript_test.go
git commit -m "fix: ConfigDir templates use bundle paths instead of source paths"
```

---

### Task 7: Fix XDG Config template to restore individual tool subdirs

The XDG template uses `{{.ConfigDir}}` (old pattern) and only logs tool names without restoring them. Since the bundler copies each auto-detected tool to `configs/xdg-config/<toolname>/`, the template must copy them back.

**Files:**
- Modify: `templates/stages/tools.sh.tmpl`
- Modify: `internal/bundler/restorescript_test.go`

**Step 1: Write the failing test**

```go
func TestGenerateRestoreScript_XDGConfigRestoresToolDirs(t *testing.T) {
	snap := &domain.Snapshot{
		Meta: newMeta(),
		XDGConfig: &domain.XDGConfigSection{
			AutoDetected: []string{"bat", "lazygit", "starship"},
		},
	}

	script, err := GenerateRestoreScript(snap)
	require.NoError(t, err)

	// Should restore each tool's config dir from bundle
	assert.Contains(t, script, `configs/xdg-config/bat`)
	assert.Contains(t, script, `cp -R "configs/xdg-config/bat/" "$HOME/.config/bat/"`)
	assert.Contains(t, script, `cp -R "configs/xdg-config/lazygit/" "$HOME/.config/lazygit/"`)
	assert.Contains(t, script, `cp -R "configs/xdg-config/starship/" "$HOME/.config/starship/"`)

	// Should NOT reference raw ConfigDir
	assert.NotContains(t, script, `{{.ConfigDir}}`)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bundler/ -run TestGenerateRestoreScript_XDGConfig -v`
Expected: FAIL

**Step 3: Fix the template**

Replace the xdg-config template in `tools.sh.tmpl`:

```
{{define "xdg-config"}}
log "Restoring XDG configs..."
mkdir -p "$HOME/.config"

{{range .AutoDetected}}
if [ -d "configs/xdg-config/{{.}}" ]; then
    log "Restoring XDG config: {{.}}"
    mkdir -p "$HOME/.config/{{.}}"
    cp -R "configs/xdg-config/{{.}}/" "$HOME/.config/{{.}}/"
fi
{{end}}
{{end}}
```

**Step 4: Run tests**

Run: `go test ./internal/bundler/ -v`
Expected: All pass.

**Step 5: Commit**

```bash
git add templates/stages/tools.sh.tmpl internal/bundler/restorescript_test.go
git commit -m "fix: XDG config template restores individual tool subdirs from bundle"
```

---

### Task 8: Add font bundling to collectConfigFiles

Custom fonts in `FontsSection.CustomFonts` are never copied to the bundle. The fonts template expects them at `configs/fonts/<name>`.

**Files:**
- Modify: `internal/bundler/dmg.go`
- Modify: `internal/bundler/dmg_test.go`

**Step 1: Write the failing test**

```go
func TestCollectConfigFiles_IncludesFonts(t *testing.T) {
	snap := &domain.Snapshot{
		Fonts: &domain.FontsSection{
			CustomFonts: []domain.Font{
				{Name: "FiraCode-Regular.ttf", Source: "Library/Fonts/FiraCode-Regular.ttf"},
				{Name: "JetBrainsMono.ttf", Source: "Library/Fonts/JetBrainsMono.ttf"},
			},
		},
	}

	files := collectConfigFiles(snap)
	require.Len(t, files, 2)
	assert.Equal(t, "Library/Fonts/FiraCode-Regular.ttf", files[0].Source)
	assert.Equal(t, "configs/fonts/FiraCode-Regular.ttf", files[0].BundlePath)
	assert.Equal(t, "Library/Fonts/JetBrainsMono.ttf", files[1].Source)
	assert.Equal(t, "configs/fonts/JetBrainsMono.ttf", files[1].BundlePath)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/bundler/ -run TestCollectConfigFiles_IncludesFonts -v`
Expected: FAIL — fonts not in collectConfigFiles.

**Step 3: Add fonts to collectConfigFiles**

In `internal/bundler/dmg.go`, add after the Raycast/AITools block in `collectConfigFiles()`:

```go
// Custom fonts
if s := snapshot.Fonts; s != nil {
    for _, font := range s.CustomFonts {
        if font.Source != "" {
            files = append(files, wrap(font.Source, "fonts"))
        }
    }
}
```

**Step 4: Run tests**

Run: `go test ./internal/bundler/ -v`
Expected: All pass.

**Step 5: Commit**

```bash
git add internal/bundler/dmg.go internal/bundler/dmg_test.go
git commit -m "fix: bundle custom fonts into configs/fonts/"
```

---

### Task 9: Run full test suite and verify end-to-end

**Step 1: Run all tests**

```bash
go test ./... -count=1
```

Expected: All pass.

**Step 2: Build and test a fresh snapshot + DMG**

```bash
go build -o machinist ./cmd/machinist
./machinist scan -o /tmp/test-snapshot.toml
./machinist dmg /tmp/test-snapshot.toml -o /tmp/test-machinist.dmg
```

**Step 3: Verify the generated install.command**

Mount the DMG and inspect the install.command:
```bash
hdiutil attach /tmp/test-machinist.dmg
cat "/Volumes/Machinist Restore/install.command" | head -80
```

Verify:
- `eval "$(/opt/homebrew/bin/brew shellenv)"` appears after Homebrew install
- SSH stage appears before Git Repos stage
- `configs/gpg/` appears (not `configs/gnupg/`)
- `configs/docker/config.json` appears (not `.docker/config.json`)
- `configs/github-cli/` appears (not `.config/gh`)
- `configs/xdg-config/bat/` etc. appear

```bash
hdiutil detach "/Volumes/Machinist Restore"
```

**Step 4: Final commit (if any remaining changes)**

```bash
git add -A
git commit -m "chore: verify end-to-end DMG restore template correctness"
```
