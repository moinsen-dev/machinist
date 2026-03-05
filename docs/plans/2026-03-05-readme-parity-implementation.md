# README-Parity Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use moinsenpowers:executing-plans to implement this plan task-by-task.

**Goal:** Bring machinist implementation to full README parity — all promised scanners, CLI commands, security pipeline, and restore templates.

**Architecture:** Extend the existing Scanner interface pattern. Each scanner is a struct with `CommandRunner`, implements `Name()/Description()/Category()/Scan()`. Results flow through `ScanResult` → `applyResult()` → `Snapshot`. Restore templates are Go `text/template` fragments included by `install.command.tmpl`.

**Tech Stack:** Go 1.25, cobra CLI, BurntSushi/toml, filippo.io/age, text/template

---

## Scanner Pattern Reference

Every new scanner follows this exact pattern. Read `internal/scanner/runtimes/rust.go` and `internal/scanner/runtimes/rust_test.go` as the canonical example.

**Scanner file** (`internal/scanner/<category>/<name>.go`):
```go
package <category>

import (
    "context"
    "github.com/moinsen-dev/machinist/internal/domain"
    "github.com/moinsen-dev/machinist/internal/scanner"
    "github.com/moinsen-dev/machinist/internal/util"
)

type XxxScanner struct {
    cmd     util.CommandRunner
    homeDir string // only if needed
}

func NewXxxScanner(homeDir string, cmd util.CommandRunner) *XxxScanner {
    return &XxxScanner{homeDir: homeDir, cmd: cmd}
}

func (s *XxxScanner) Name() string        { return "xxx" }
func (s *XxxScanner) Description() string  { return "Scans ..." }
func (s *XxxScanner) Category() string     { return "<category>" }

func (s *XxxScanner) Scan(ctx context.Context) (*scanner.ScanResult, error) {
    result := &scanner.ScanResult{ScannerName: s.Name()}

    if !s.cmd.IsInstalled(ctx, "xxx") {
        return result, nil // not installed = skip
    }

    section := &domain.XxxSection{}
    // ... populate section ...
    result.Xxx = section
    return result, nil
}
```

**Test file** (`internal/scanner/<category>/<name>_test.go`):
```go
func TestXxxScanner_Scan_HappyPath(t *testing.T) {
    mock := &util.MockCommandRunner{
        Responses: map[string]util.MockResponse{
            "xxx": {Output: "xxx 1.0"}, // IsInstalled check
            "xxx --version": {Output: "1.0.0"},
        },
    }
    s := NewXxxScanner(t.TempDir(), mock)
    result, err := s.Scan(context.Background())
    require.NoError(t, err)
    require.NotNil(t, result.Xxx)
    assert.Equal(t, "1.0.0", result.Xxx.Version)
}

func TestXxxScanner_Scan_NotInstalled(t *testing.T) {
    mock := &util.MockCommandRunner{Responses: map[string]util.MockResponse{}}
    s := NewXxxScanner(t.TempDir(), mock)
    result, err := s.Scan(context.Background())
    require.NoError(t, err)
    assert.Nil(t, result.Xxx)
}
```

**Wiring checklist** (for every new scanner):
1. Add field to `scanner.ScanResult` struct in `internal/scanner/scanner.go`
2. Add case to `applyResult()` in `internal/scanner/scanner.go`
3. Add `StageCount()` case in `internal/domain/snapshot.go`
4. Register in `cmd/machinist/registry.go`
5. Add restore stage template in `templates/stages/<name>.sh.tmpl`
6. Add `{{if .Section}}` block in `templates/install.command.tmpl`

---

## Batch 0: Infrastructure

### Task 0.1: Deduplicate applyResult

**Files:**
- Modify: `internal/scanner/scanner.go` — export `ApplyResult()` (capital A)
- Modify: `cmd/machinist/registry.go` — delete `applyResultToSnapshot()`, use `scanner.ApplyResult()`
- Modify: `cmd/machinist/scan.go` — call `scanner.ApplyResult()` instead
- Modify: `internal/mcp/server.go` — call `scanner.ApplyResult()` instead of local copy

**Step 1: Rename `applyResult` → `ApplyResult` in scanner.go**
Make it exported. Keep existing field mappings.

**Step 2: Delete `applyResultToSnapshot` from registry.go**
Replace all call sites with `scanner.ApplyResult(snap, result)`.

**Step 3: Update scan.go and server.go**
Both should import and call `scanner.ApplyResult()`.

**Step 4: Run tests**
Run: `go test ./...`
Expected: All pass (this is a pure refactor)

**Step 5: Commit**
```
refactor: deduplicate applyResult into single exported function
```

### Task 0.2: Expand ScanResult with all missing fields

**Files:**
- Modify: `internal/scanner/scanner.go` — add ~30 new fields to `ScanResult`
- Modify: `internal/scanner/scanner.go` — expand `ApplyResult()` for all new fields

**Step 1: Add fields to ScanResult**
Add one field per missing Snapshot section. Fields to add:
```go
Java          *domain.JavaSection
Flutter       *domain.FlutterSection
Go            *domain.GoSection
Asdf          *domain.AsdfSection
Deno          *domain.DenoSection        // new type
Bun           *domain.BunSection         // new type
Ruby          *domain.RubySection        // new type
Terminal      *domain.TerminalSection
Tmux          *domain.TmuxSection
GitConfig     *domain.GitSection         // rename existing Git to GitConfig for clarity
GitHubCLI     *domain.GitHubCLISection
Neovim        *domain.NeovimSection
JetBrains     *domain.JetBrainsSection
Xcode         *domain.XcodeSection
AWS           *domain.AWSSection
Kubernetes    *domain.KubernetesSection
Terraform     *domain.TerraformSection
Vercel        *domain.VercelSection
GCP           *domain.GCPSection         // new type
Azure         *domain.AzureSection       // new type
Flyio         *domain.FlyioSection       // new type
Firebase      *domain.FirebaseSection    // new type
Cloudflare    *domain.CloudflareSection  // new type
Locale        *domain.LocaleSection
LoginItems    *domain.LoginItemsSection
HostsFile     *domain.HostsFileSection
Network       *domain.NetworkSection
SSH           *domain.SSHSection
GPG           *domain.GPGSection
Raycast       *domain.RaycastSection
Alfred        *domain.AlfredSection      // new type
Karabiner     *domain.KarabinerSection
Rectangle     *domain.RectangleSection
BetterTouchTool *domain.BetterTouchToolSection // new type
OnePassword   *domain.OnePasswordSection       // new type
Browser       *domain.BrowserSection
AITools       *domain.AIToolsSection
APITools      *domain.APIToolsSection
Databases     *domain.DatabasesSection
Registries    *domain.RegistriesSection
XDGConfig     *domain.XDGConfigSection
EnvFiles      *domain.EnvFilesSection
```

**Step 2: Expand ApplyResult()**
Add `if result.Xxx != nil { snap.Xxx = result.Xxx }` for every new field.

**Step 3: Run tests**
Run: `go test ./...`

**Step 4: Commit**
```
feat: expand ScanResult to support all snapshot sections
```

### Task 0.3: Add new domain types

**Files:**
- Modify: `internal/domain/snapshot.go` — add new section types and Snapshot fields
- Modify: `internal/domain/types.go` — add new value types if needed

**Step 1: Add new section types to snapshot.go**
```go
type DenoSection struct {
    Version        string    `toml:"version,omitempty"`
    GlobalPackages []Package `toml:"global_packages,omitempty"`
}

type BunSection struct {
    Version        string    `toml:"version,omitempty"`
    GlobalPackages []Package `toml:"global_packages,omitempty"`
}

type RubySection struct {
    Manager        string    `toml:"manager,omitempty"`
    Versions       []string  `toml:"versions,omitempty"`
    DefaultVersion string    `toml:"default_version,omitempty"`
    GlobalGems     []Package `toml:"global_gems,omitempty"`
}

type GCPSection struct {
    ConfigDir string `toml:"config_dir,omitempty"`
}

type AzureSection struct {
    ConfigDir string `toml:"config_dir,omitempty"`
}

type FlyioSection struct {
    ConfigFile string `toml:"config_file,omitempty"`
}

type FirebaseSection struct {
    ConfigDir string `toml:"config_dir,omitempty"`
}

type CloudflareSection struct {
    ConfigDir string `toml:"config_dir,omitempty"`
}

type AlfredSection struct {
    ConfigDir string `toml:"config_dir,omitempty"`
}

type BetterTouchToolSection struct {
    ConfigFile string `toml:"config_file,omitempty"`
}

type OnePasswordSection struct {
    ConfigDir string `toml:"config_dir,omitempty"`
}
```

**Step 2: Add fields to Snapshot struct**
```go
Deno            *DenoSection            `toml:"deno,omitempty"`
Bun             *BunSection             `toml:"bun,omitempty"`
Ruby            *RubySection            `toml:"ruby,omitempty"`
GCP             *GCPSection             `toml:"gcp,omitempty"`
Azure           *AzureSection           `toml:"azure,omitempty"`
Flyio           *FlyioSection           `toml:"flyio,omitempty"`
Firebase        *FirebaseSection        `toml:"firebase,omitempty"`
Cloudflare      *CloudflareSection      `toml:"cloudflare,omitempty"`
Alfred          *AlfredSection          `toml:"alfred,omitempty"`
BetterTouchTool *BetterTouchToolSection `toml:"bettertouchtool,omitempty"`
OnePassword     *OnePasswordSection     `toml:"onepassword,omitempty"`
```

**Step 3: Expand StageCount()**
Add `if s.Xxx != nil { count++ }` for every new section.

**Step 4: Run tests**
Run: `go test ./...`

**Step 5: Commit**
```
feat: add domain types for all README-promised sections
```

### Task 0.4: CLI fixes

**Files:**
- Modify: `cmd/machinist/snapshot.go` — add `--dry-run` flag
- Modify: `cmd/machinist/scan.go` — add `--search-paths` flag
- Modify: `cmd/machinist/compose.go` — add `--from-file` flag
- Modify: `cmd/machinist/restore.go` — make positional arg optional (default `manifest.toml`)
- Modify: `cmd/machinist/list.go` — add `list scanners` and `list profiles` subcommands
- Modify: `README.md` — update Usage section to match actual CLI

**Step 1: snapshot --dry-run**
Add `snapshotDryRun` bool flag. When set: scan, marshal to TOML, print to stdout, don't write file.

**Step 2: scan --search-paths**
Add `scanSearchPaths` string flag. If provided and scanner is `git-repos`, split by comma and pass to `NewGitReposScanner()`. This requires the registry to be parameterized.

**Step 3: compose --from-file**
Add `composeFromFile` string flag. When set: `domain.ReadManifest(composeFromFile)` as base instead of profile. Positional arg becomes optional when `--from-file` is used.

**Step 4: restore default arg**
Change `cobra.ExactArgs(1)` to `cobra.MaximumNArgs(1)`. Default to `manifest.toml` in CWD.

**Step 5: list subcommands**
Add `listScannersCmd` and `listProfilesCmd` as children of `listCmd`. Keep `listCmd` showing both.

**Step 6: Update README.md Usage section**
- `compose <profile>` (not `--from`)
- Add `compose --from-file <manifest.toml>`
- Add `dmg` command
- `restore [manifest.toml]` (optional arg)
- `list` / `list scanners` / `list profiles`
- Document `--yes` for restore

**Step 7: Run tests**
Run: `go test ./cmd/machinist/...`

**Step 8: Commit**
```
fix: align CLI commands and flags with README documentation
```

### Task 0.5: Restore template infrastructure

**Files:**
- Modify: `templates/install.command.tmpl` — add `{{if .Section}}` blocks for all new sections
- Create: stage template stubs in `templates/stages/` for new categories

**Step 1: Add template blocks to install.command.tmpl**
For each new section, add:
```
{{if .Go}}
do_go() {
{{template "go-runtime" .Go}}
}
run_stage "Go" do_go
{{end}}
```

Stage template files to create (can be stubs initially, filled in per batch):
- `templates/stages/runtimes_extra.sh.tmpl` — go, java, flutter, deno, bun, ruby, asdf
- `templates/stages/security.sh.tmpl` — ssh, gpg
- `templates/stages/editors_extra.sh.tmpl` — jetbrains, neovim, xcode
- `templates/stages/terminal.sh.tmpl` — terminal emulators, tmux
- `templates/stages/cloud.sh.tmpl` — docker, aws, gcp, azure, k8s, terraform, vercel, flyio, firebase, cloudflare
- `templates/stages/macos_extra.sh.tmpl` — locale, login-items, hosts, network
- `templates/stages/tools.sh.tmpl` — raycast, alfred, karabiner, rectangle, btt, 1password, databases, registries, browser, ai-tools, api-tools, xdg, env-files

**Step 2: Create stub templates**
Each defines a named template that logs "Restoring <name>..." as a placeholder.

**Step 3: Run tests**
Run: `go test ./internal/bundler/...`

**Step 4: Commit**
```
feat: add restore template slots for all new scanner sections
```

### Task 0.6: Dynamic checklist

**Files:**
- Modify: `templates/checklist.md.tmpl`

**Step 1: Replace static content with conditional blocks**

```
# Post-Restore Checklist

Restored from: **{{.Meta.SourceHostname}}** ({{.Meta.SourceOSVersion}}, {{.Meta.SourceArch}})

The following items need manual attention:

## macOS Permissions (TCC)
- [ ] Grant Terminal Full Disk Access (System Preferences -> Privacy & Security)
{{if .Raycast}}- [ ] Grant Raycast accessibility permissions{{end}}
{{if .Rectangle}}- [ ] Grant Rectangle accessibility permissions{{end}}
{{if .Karabiner}}- [ ] Grant Karabiner-Elements input monitoring{{end}}

{{if .Browser}}
## Browser
- [ ] Sign in to {{.Browser.Default}} and sync extensions
- [ ] Import bookmarks if needed
{{end}}

## Accounts & Credentials
- [ ] Sign in to Mac App Store
{{if .Xcode}}- [ ] Sign in to Xcode with Apple ID{{end}}
{{if .Docker}}- [ ] Sign in to Docker Desktop{{end}}
{{if .GCP}}- [ ] Run `gcloud auth login`{{end}}
{{if .Azure}}- [ ] Run `az login`{{end}}
{{if .OnePassword}}- [ ] Run `op signin`{{end}}

{{if or .SSH .GPG}}
## Git & SSH
{{if .SSH}}- [ ] Test SSH key authentication: `ssh -T git@github.com`{{end}}
{{if .GPG}}- [ ] Verify git signing works: `git log --show-signature`{{end}}
{{end}}

{{if .Network}}
## Network
- [ ] Set up VPN connections manually
- [ ] Verify DNS settings
{{end}}
```

**Step 2: Run tests**
Run: `go test ./internal/bundler/...`

**Step 3: Commit**
```
feat: make post-restore checklist dynamic based on snapshot content
```

---

## Batch 1: Runtimes (7 scanners)

For each scanner below, follow the Scanner Pattern Reference. Files per scanner:
- Create: `internal/scanner/runtimes/<name>.go`
- Create: `internal/scanner/runtimes/<name>_test.go`
- Fill in restore template in `templates/stages/runtimes_extra.sh.tmpl`

### Task 1.1: Go scanner

**Scanner:** `internal/scanner/runtimes/go.go`
- `IsInstalled(ctx, "go")` → skip if not found
- `go version` → parse version
- `go env GOPATH` → get GOPATH
- List binaries in `$GOPATH/bin/` (just names, these represent `go install`ed tools)
- Fill `GoSection{Version, GlobalPackages}`

**Restore template:**
```
{{define "go-runtime"}}
# Go should be installed via Homebrew or manually
{{range .GlobalPackages}}
log "Installing Go tool {{.Name}}"
go install "{{.Name}}@latest" || true
{{end}}
{{end}}
```

### Task 1.2: Java/SDKMAN scanner

**Scanner:** `internal/scanner/runtimes/java.go`
- Check for `sdk` command (SDKMAN) first via `bash -c "source ~/.sdkman/bin/sdkman-init.sh && sdk version"`
- If SDKMAN: `sdk list java` to find installed versions, parse `sdk current java` for default
- Fallback: `java -version` (stderr!) for system Java, `echo $JAVA_HOME`
- Fill `JavaSection{Manager, Versions, DefaultVersion, JavaHome}`

**Restore template:**
```
{{define "java"}}
{{if eq .Manager "sdkman"}}
if [ ! -d "$HOME/.sdkman" ]; then
    curl -s "https://get.sdkman.io" | bash
fi
source "$HOME/.sdkman/bin/sdkman-init.sh"
{{range .Versions}}
sdk install java "{{.}}" || true
{{end}}
{{if .DefaultVersion}}sdk default java "{{.DefaultVersion}}"{{end}}
{{end}}
{{end}}
```

### Task 1.3: Flutter scanner

**Scanner:** `internal/scanner/runtimes/flutter.go`
- `IsInstalled(ctx, "flutter")` → skip if not found
- `flutter --version` → parse channel and version
- `dart pub global list` → parse global packages
- Fill `FlutterSection{Channel, Version, DartGlobalPackages}`

**Restore template:**
```
{{define "flutter"}}
{{range .DartGlobalPackages}}
log "Activating Dart package {{.}}"
dart pub global activate "{{.}}" || true
{{end}}
{{end}}
```

### Task 1.4: Deno scanner

**Scanner:** `internal/scanner/runtimes/deno.go`
- `IsInstalled(ctx, "deno")` → skip if not found
- `deno --version` → parse version (first line)
- List `~/.deno/bin/` for globally installed scripts
- Fill `DenoSection{Version, GlobalPackages}`

**Restore template:**
```
{{define "deno"}}
if ! command -v deno &>/dev/null; then
    curl -fsSL https://deno.land/install.sh | sh
fi
{{range .GlobalPackages}}
log "Installing Deno package {{.Name}}"
deno install "{{.Name}}" || true
{{end}}
{{end}}
```

### Task 1.5: Bun scanner

**Scanner:** `internal/scanner/runtimes/bun.go`
- `IsInstalled(ctx, "bun")` → skip if not found
- `bun --version` → version
- `bun pm ls -g` → parse global packages (lines like `package@version`)
- Fill `BunSection{Version, GlobalPackages}`

**Restore template:**
```
{{define "bun"}}
if ! command -v bun &>/dev/null; then
    curl -fsSL https://bun.sh/install | bash
fi
{{range .GlobalPackages}}
log "Installing Bun package {{.Name}}"
bun install -g "{{.Name}}" || true
{{end}}
{{end}}
```

### Task 1.6: Ruby scanner

**Scanner:** `internal/scanner/runtimes/ruby.go`
- Check `rbenv` → `rbenv versions`, `rbenv global` for default
- Check `rvm` → `rvm list`, `rvm use --default`
- Fallback: `ruby --version`
- `gem list --no-versions` → global gems (skip default gems)
- Fill `RubySection{Manager, Versions, DefaultVersion, GlobalGems}`

**Restore template:**
```
{{define "ruby"}}
{{if eq .Manager "rbenv"}}
if ! command -v rbenv &>/dev/null; then
    brew install rbenv ruby-build
fi
{{range .Versions}}
rbenv install -s "{{.}}"
{{end}}
{{if .DefaultVersion}}rbenv global "{{.DefaultVersion}}"{{end}}
{{end}}
{{range .GlobalGems}}
gem install "{{.Name}}" || true
{{end}}
{{end}}
```

### Task 1.7: asdf/mise scanner

**Scanner:** `internal/scanner/runtimes/asdf.go`
- Check `asdf` or `mise`
- `asdf plugin list` → plugins
- For each plugin: `asdf list <plugin>` → versions
- `~/.tool-versions` → capture as config file
- Fill `AsdfSection{Plugins, ToolVersionsFile}`

**Restore template:**
```
{{define "asdf"}}
{{range .Plugins}}
log "Adding asdf plugin {{.Name}}"
asdf plugin add "{{.Name}}" || true
{{range .Versions}}
asdf install "{{$.Name}}" "{{.}}" || true
{{end}}
{{end}}
{{if .ToolVersionsFile}}
cp "configs/.tool-versions" "$HOME/.tool-versions"
{{end}}
{{end}}
```

**After Batch 1:** Register all 7 scanners in `registry.go`. Run `go test ./...`. Commit.

---

## Batch 2: Git & Security (4 scanners + security wiring)

### Task 2.1: git-config scanner

**Scanner:** `internal/scanner/git/config.go`
- Scan `~/.gitconfig` or `~/.config/git/config` as ConfigFile
- `~/.gitignore_global` or `git config --global core.excludesfile`
- `git config --global user.signingkey` + `git config --global gpg.format` → signing_method
- `git config --global credential.helper` → credential_helper
- `git config --global init.templateDir` → template_dir
- Fill `GitSection`

### Task 2.2: github-cli scanner

**Scanner:** `internal/scanner/git/githubcli.go`
- `IsInstalled(ctx, "gh")` → skip if not found
- `~/.config/gh/` as config dir
- `gh extension list` → parse extension names
- Fill `GitHubCLISection`

### Task 2.3: ssh scanner

**Scanner:** `internal/scanner/security/ssh.go`
- Check `~/.ssh/` exists
- List key files (`id_rsa`, `id_ed25519`, etc.) — only names, not content
- `~/.ssh/config` as ConfigFile
- `~/.ssh/known_hosts` path
- Set `Encrypted = true` (these will be age-encrypted during bundling)
- Fill `SSHSection`

### Task 2.4: gpg scanner

**Scanner:** `internal/scanner/security/gpg.go`
- `IsInstalled(ctx, "gpg")` → skip if not found
- `gpg --list-keys --keyid-format long` → parse key IDs
- `~/.gnupg/gpg.conf`, `~/.gnupg/gpg-agent.conf` as ConfigFiles
- Set `Encrypted = true`
- Fill `GPGSection`

### Task 2.5: Wire security into bundler

**Files:**
- Modify: `internal/bundler/dmg.go` — import `internal/security`, encrypt files with `Encrypted=true`
- Modify: `cmd/machinist/snapshot.go` or `cmd/machinist/dmg.go` — prompt for age passphrase when secrets detected

**Logic in `PrepareBundleDir`:**
```go
// After copying config files, encrypt any with Encrypted=true
for _, cf := range configFiles {
    if cf.Encrypted {
        encPath := filepath.Join(bundleDir, cf.BundlePath)
        if err := security.EncryptFile(encPath, encPath+".age", passphrase); err != nil {
            return fmt.Errorf("encrypt %s: %w", cf.Source, err)
        }
        os.Remove(encPath) // remove plaintext
    }
}
```

**After Batch 2:** Register scanners, run tests, commit.

---

## Batch 3: Editors (3 scanners + 1 fix)

### Task 3.1: JetBrains scanner

**Scanner:** `internal/scanner/editors/jetbrains.go`
- Walk `~/Library/Application Support/JetBrains/` for IDE directories
- Known IDE prefixes: IntelliJIdea, PyCharm, WebStorm, GoLand, AndroidStudio, PhpStorm, CLion, Rider, DataGrip, RubyMine
- For each found: record name and settings export path
- Fill `JetBrainsSection{IDEs}`

### Task 3.2: Neovim scanner

**Scanner:** `internal/scanner/editors/neovim.go`
- `IsInstalled(ctx, "nvim")` → skip if not found
- Check `~/.config/nvim/` exists
- Detect plugin manager: `lazy.nvim` (check `~/.local/share/nvim/lazy/`), `packer` (check `packer_compiled.lua`), `vim-plug` (check `plug.vim`)
- Fill `NeovimSection{ConfigDir, PluginManager}`

### Task 3.3: Xcode scanner

**Scanner:** `internal/scanner/editors/xcode.go`
- `IsInstalled(ctx, "xcode-select")` → skip if not found
- `xcrun simctl list devices -j` → parse JSON for simulator names
- Xcode pref files in `~/Library/Preferences/com.apple.dt.Xcode.plist`
- Fill `XcodeSection{Simulators, ConfigFiles}`

### Task 3.4: Cursor scanner fix

**Files:**
- Modify: `internal/scanner/editors/vscode.go` — add `keybindings.json` and `snippets/` to CursorScanner

**After Batch 3:** Register, test, commit.

---

## Batch 4: Shell & Terminal (3 scanners + 1 fix)

### Task 4.1: Terminal emulator scanner

**Scanner:** `internal/scanner/shell/terminal.go`
- Detect terminal app in use. Check for config files:
  - iTerm2: `~/Library/Preferences/com.googlecode.iterm2.plist`
  - Warp: `~/.warp/`
  - Alacritty: `~/.config/alacritty/alacritty.toml` or `.yml`
  - WezTerm: `~/.config/wezterm/wezterm.lua`
- First match wins for `App` field
- Capture config files
- Fill `TerminalSection{App, ConfigFiles}`

### Task 4.2: Tmux scanner

**Scanner:** `internal/scanner/shell/tmux.go`
- `IsInstalled(ctx, "tmux")` → skip if not found
- Check `~/.tmux.conf` or `~/.config/tmux/tmux.conf`
- Parse TPM plugins: grep for `set -g @plugin` lines
- Fill `TmuxSection{ConfigFiles, TPMPlugins}`

### Task 4.3: oh-my-zsh custom plugins fix

**Files:**
- Modify: `internal/scanner/shell/config.go` — add plugin directory scanning
- Read `~/.oh-my-zsh/custom/plugins/`, list directory names
- Populate `ShellSection.OhMyZshCustomPlugins`

### Task 4.4: direnv integration

**Files:**
- Modify: `internal/scanner/shell/config.go`
- Check `direnv` installed, add `~/.direnvrc` or `~/.config/direnv/direnvrc` to config files

**After Batch 4:** Register, test, commit.

---

## Batch 5: Containers & Cloud (11 scanners)

### Task 5.1: Docker scanner

**Scanner:** `internal/scanner/cloud/docker.go`
- `IsInstalled(ctx, "docker")` → skip if not found
- `~/.docker/config.json` as config file
- Detect runtime: check for `colima status`, `orbctl version`, `podman --version`, else assume Docker Desktop
- `docker images --format '{{.Repository}}:{{.Tag}}'` → frequently used images (limit to top 20, skip `<none>`)
- Fill `DockerSection{ConfigFile, FrequentlyUsedImages, Runtime}`

### Task 5.2: AWS scanner

**Scanner:** `internal/scanner/cloud/aws.go`
- `IsInstalled(ctx, "aws")` → skip if not found
- `~/.aws/config` path
- `aws configure list-profiles` → profile names
- Fill `AWSSection{ConfigFile, Profiles}`

### Task 5.3: Kubernetes scanner

**Scanner:** `internal/scanner/cloud/kubernetes.go`
- `IsInstalled(ctx, "kubectl")` → skip if not found
- `~/.kube/config` path
- `kubectl config get-contexts -o name` → context names
- Fill `KubernetesSection{ConfigFile, Contexts}`

### Task 5.4: Terraform scanner

**Scanner:** `internal/scanner/cloud/terraform.go`
- `IsInstalled(ctx, "terraform")` → skip if not found
- `~/.terraformrc` path
- Fill `TerraformSection{ConfigFile}`

### Task 5.5: Vercel scanner

**Scanner:** `internal/scanner/cloud/vercel.go`
- `IsInstalled(ctx, "vercel")` → skip if not found
- `~/.vercel/` as config dir
- Fill `VercelSection{ConfigDir}`

### Task 5.6: GCP scanner

**Scanner:** `internal/scanner/cloud/gcp.go`
- `IsInstalled(ctx, "gcloud")` → skip if not found
- `~/.config/gcloud/` as config dir
- Fill `GCPSection{ConfigDir}`

### Task 5.7: Azure scanner

**Scanner:** `internal/scanner/cloud/azure.go`
- `IsInstalled(ctx, "az")` → skip if not found
- `~/.azure/` as config dir
- Fill `AzureSection{ConfigDir}`

### Task 5.8: Fly.io scanner

**Scanner:** `internal/scanner/cloud/flyio.go`
- `IsInstalled(ctx, "fly")` → skip if not found
- `~/.fly/config.yml` path
- Fill `FlyioSection{ConfigFile}`

### Task 5.9: Firebase scanner

**Scanner:** `internal/scanner/cloud/firebase.go`
- `IsInstalled(ctx, "firebase")` → skip if not found
- `~/.config/configstore/firebase-tools.json` or `~/.config/firebase/`
- Fill `FirebaseSection{ConfigDir}`

### Task 5.10: Cloudflare scanner

**Scanner:** `internal/scanner/cloud/cloudflare.go`
- `IsInstalled(ctx, "wrangler")` → skip if not found
- `~/.config/.wrangler/` or `~/.wrangler/config/`
- Fill `CloudflareSection{ConfigDir}`

**After Batch 5:** Register all, test, commit.

---

## Batch 6: macOS System (5 scanners + 4 fixes)

### Task 6.1: macOS Defaults — extend for Trackpad, Mission Control, Spotlight, Menu Bar

**Files:**
- Modify: `internal/scanner/system/macos_defaults.go`
- Modify: `internal/scanner/system/macos_defaults_test.go`

**Trackpad:**
- `defaults read com.apple.AppleMultitouchTrackpad Clicking` → tap_to_click
- `defaults read NSGlobalDomain com.apple.trackpad.scaling` → tracking_speed

**Mission Control:**
- `defaults read com.apple.dock wvous-tl-corner` → hot corners (tl, tr, bl, br)
- Map integer values to action names (2=Mission Control, 4=Desktop, 5=Screensaver, etc.)

**Spotlight:**
- `defaults read com.apple.Spotlight orderedItems` → parse for disabled categories

**Menu Bar:**
- `defaults read com.apple.menuextra.clock DateFormat` → clock format
- `defaults read com.apple.menuextra.battery ShowPercent` → show battery percentage

### Task 6.2: Locale scanner

**Scanner:** `internal/scanner/system/locale.go`
- `defaults read NSGlobalDomain AppleLanguages` → language (first entry)
- `defaults read NSGlobalDomain AppleLocale` → region
- `systemsetup -gettimezone` → timezone (parse after "Time Zone: ")
- `scutil --get ComputerName` → computer_name
- `scutil --get LocalHostName` → local_hostname
- Fill `LocaleSection`

### Task 6.3: Login Items scanner

**Scanner:** `internal/scanner/system/loginitems.go`
- `osascript -e 'tell application "System Events" to get the name of every login item'` → parse comma-separated
- Fill `LoginItemsSection{Apps}`

### Task 6.4: Hosts File scanner

**Scanner:** `internal/scanner/system/hosts.go`
- Read `/etc/hosts`
- Filter out standard entries (lines starting with `127.0.0.1 localhost`, `255.255.255.255`, `::1`, `fe80::1`)
- Parse remaining as `HostEntry{IP, Hostnames}`
- Fill `HostsFileSection{CustomEntries}`

### Task 6.5: Network scanner

**Scanner:** `internal/scanner/system/network.go`
- `networksetup -listpreferredwirelessnetworks en0` → Wi-Fi names (skip header line)
- `networksetup -getdnsservers Wi-Fi` → DNS servers
- `scutil --nc list` → VPN config names
- If `tailscale version` succeeds, note Tailscale is installed
- Fill `NetworkSection{PreferredWifi, DNS, VPNConfigs}`

**After Batch 6:** Register, update macos_defaults restore template, test, commit.

---

## Batch 7: Tools & Config (14 scanners)

### Task 7.1: Raycast scanner

**Scanner:** `internal/scanner/tools/raycast.go`
- Check `~/Library/Application Support/com.raycast.macos/` exists
- Or check for Raycast export file
- Fill `RaycastSection{ExportFile}`

### Task 7.2: Alfred scanner

**Scanner:** `internal/scanner/tools/alfred.go`
- Check `~/Library/Application Support/Alfred/` exists
- Fill `AlfredSection{ConfigDir}`

### Task 7.3: Karabiner scanner

**Scanner:** `internal/scanner/tools/karabiner.go`
- Check `~/.config/karabiner/karabiner.json` exists
- Fill `KarabinerSection{ConfigDir}`

### Task 7.4: Rectangle scanner

**Scanner:** `internal/scanner/tools/rectangle.go`
- Check `~/Library/Preferences/com.knewton.Rectangle.plist` exists
- Fill `RectangleSection{ConfigFile}`

### Task 7.5: BetterTouchTool scanner

**Scanner:** `internal/scanner/tools/bettertouchtool.go`
- Check `~/Library/Application Support/BetterTouchTool/` exists
- Fill `BetterTouchToolSection{ConfigFile}`

### Task 7.6: 1Password CLI scanner

**Scanner:** `internal/scanner/tools/onepassword.go`
- `IsInstalled(ctx, "op")` → skip if not found
- `~/.config/op/` as config dir
- Fill `OnePasswordSection{ConfigDir}`

### Task 7.7: Databases scanner

**Scanner:** `internal/scanner/tools/databases.go`
- Check for: `~/.pgpass`, `~/.my.cnf`, `~/Library/Application Support/com.tinyapp.TablePlus/`, `~/.dbeaver4/`
- Collect as ConfigFiles, mark `.pgpass` and `.my.cnf` as Sensitive
- Fill `DatabasesSection{ConfigFiles}`

### Task 7.8: Registries scanner

**Scanner:** `internal/scanner/tools/registries.go`
- Check for: `~/.npmrc`, `~/.pip/pip.conf` or `~/.config/pip/pip.conf`, `~/.cargo/config.toml`, `~/.gemrc`, `~/.cocoapods/config.yaml`
- Collect as ConfigFiles, mark all as Sensitive (may contain auth tokens)
- Fill `RegistriesSection{ConfigFiles}`

### Task 7.9: Browser scanner

**Scanner:** `internal/scanner/tools/browser.go`
- `defaults read com.apple.LaunchServices/com.apple.launchservices.secure LSHandlers` → find HTTP handler bundle ID
- Map bundle ID to name (com.google.Chrome → Chrome, org.mozilla.firefox → Firefox, etc.)
- `ExtensionsChecklist` = "Sign in to browser to sync extensions"
- Fill `BrowserSection{Default, ExtensionsChecklist}`

### Task 7.10: AI Tools scanner

**Scanner:** `internal/scanner/tools/aitools.go`
- Check `~/.claude/` exists → ClaudeCodeConfig path
- `IsInstalled(ctx, "ollama")` → `ollama list` → parse model names
- Fill `AIToolsSection{ClaudeCodeConfig, OllamaModels}`

### Task 7.11: API Tools scanner

**Scanner:** `internal/scanner/tools/apitools.go`
- Check for: `~/Library/Application Support/Postman/`, `~/Library/Application Support/Insomnia/`
- `IsInstalled(ctx, "ngrok")` → note installed
- `IsInstalled(ctx, "mkcert")` → set `Mkcert = true`
- Fill `APIToolsSection{ConfigFiles, Mkcert}`

### Task 7.12: XDG Config scanner

**Scanner:** `internal/scanner/tools/xdgconfig.go`
- Walk `~/.config/` for known tool directories: `bat`, `lazygit`, `htop`, `btop`, `aerospace`, `gh`, `starship.toml`
- Record found directories/files as `AutoDetected`
- `ConfigDir` = `~/.config`
- Fill `XDGConfigSection{AutoDetected, CustomPaths, ConfigDir}`

### Task 7.13: Env Files scanner

**Scanner:** `internal/scanner/tools/envfiles.go`
- Walk common workspace dirs (`~/Code`, `~/Projects`, `~/work`, `~/Developer`)
- Find `.env`, `.env.local`, `.env.production` files (max depth 3)
- Record as EnvFile entries, all Encrypted
- Fill `EnvFilesSection{Encrypted: true, Files}`

### Task 7.14: Tailscale (integrated into Network scanner)

Already handled in Task 6.5. No separate scanner needed.

**After Batch 7:** Register all, fill in restore templates, test, commit.

---

## Batch 8: Final Polish

### Task 8.1: MCP server — fix profile loading bug

**Files:**
- Modify: `internal/mcp/server.go` — use `profiles.List()` and `profiles.Get()` from `embed.FS` instead of filesystem paths

### Task 8.2: Update README "31 stages" claim

**Files:**
- Modify: `README.md` — after all scanners are implemented, count actual `StageCount()` max and update the number

### Task 8.3: Remove "Display" from README

**Files:**
- Modify: `README.md` — remove Display from macOS System list (no reliable `defaults write` for display scaling across hardware)

### Task 8.4: Full integration test

**Files:**
- Modify: `internal/integration_test.go` — extend to cover new scanners

Run `go test ./...` — all must pass.

### Task 8.5: Final commit

```
feat: complete README-parity — all scanners, CLI fixes, security wiring
```

---

## Execution Order Summary

| Batch | Tasks | New Files | Description |
|---|---|---|---|
| 0 | 0.1–0.6 | ~10 templates | Infrastructure, CLI fixes, checklist |
| 1 | 1.1–1.7 | 14 | Runtime scanners |
| 2 | 2.1–2.5 | 8 + bundler mod | Git & Security |
| 3 | 3.1–3.4 | 6 | Editors |
| 4 | 4.1–4.4 | 4 + mods | Shell & Terminal |
| 5 | 5.1–5.10 | 20 | Containers & Cloud |
| 6 | 6.1–6.5 | 8 + mods | macOS System |
| 7 | 7.1–7.13 | 26 | Tools & Config |
| 8 | 8.1–8.5 | 0 | Polish & verify |

**Total: ~96 new files, ~25 modified files, 8 batches, ~45 tasks**
