# Split Restore Scripts Design

## Problem

The DMG bundles a single monolithic `install.command` (6000+ lines, 35 stages). If a stage fails or needs re-running, there's no way to resume or selectively re-execute without running everything from scratch.

## Solution

Split the restore into 7 numbered shell scripts grouped by dependency order, plus a thin orchestrator.

## DMG Structure

```
machinist/
├── install.command              # Orchestrator: runs 01-07 sequentially
├── 01-foundation.sh             # Homebrew, SSH, GPG, Git Config
├── 02-shell.sh                  # Shell, Terminal, tmux
├── 03-runtimes.sh               # Node, Python, Rust, Flutter, Go, Ruby, Java, Deno, Bun, asdf
├── 04-editors.sh                # VS Code, Cursor, Xcode, JetBrains, Neovim, GitHub CLI
├── 05-infrastructure.sh         # Docker, K8s, GCP, AWS, Azure, Terraform, Firebase, Cloudflare, Vercel, Fly.io
├── 06-repos.sh                  # Git Repos
├── 07-system.sh                 # macOS Defaults, Fonts, Apps, Folders, Scheduled, Locale, Login Items, Hosts, Network, Productivity, Browser, AI, API, XDG, .env, DBs, Registries
├── manifest.toml
├── configs/...
├── README.md
└── POST_RESTORE_CHECKLIST.md
```

## Group Dependency Order

```
01-foundation ──→ 02-shell ──→ 03-runtimes ──→ 04-editors
                                    │
                                    ├──→ 05-infrastructure
                                    │
                                    └──→ 06-repos
                                              │
                                              └──→ 07-system
```

- Foundation first: Homebrew is prerequisite for everything; SSH keys needed for git clones
- Shell before runtimes: PATH and shell plugins must be loaded before runtime installers run
- Runtimes before editors: VS Code extensions like rust-analyzer need the runtime
- Repos isolated: slowest step (400+ clones), most likely to need re-running
- System last: no dependencies, can run independently

## Group Contents

### 01-foundation.sh
- Homebrew (install, taps, formulae, casks, services)
- SSH Keys (age decrypt)
- GPG Keys (age decrypt)
- Git Config

### 02-shell.sh
- Shell config (.zshrc, .zshenv, oh-my-zsh, plugins)
- Terminal emulator configs (iTerm2, Warp, Ghostty)
- tmux

### 03-runtimes.sh
- Node.js (nvm/fnm + global packages)
- Python (pyenv/uv + global packages)
- Rust (rustup + components + cargo packages)
- Flutter/Dart (channel + global packages)
- Go (version + global packages)
- Ruby
- Java/SDKMAN
- Deno
- Bun
- asdf/mise

### 04-editors.sh
- VS Code (settings, keybindings, extensions)
- Cursor
- Xcode (plist, simulators)
- JetBrains IDEs
- Neovim
- GitHub CLI (config, extensions)

### 05-infrastructure.sh
- Docker
- Kubernetes
- GCP
- AWS
- Azure
- Terraform
- Firebase
- Cloudflare
- Vercel
- Fly.io

### 06-repos.sh
- Git repositories (all clones)

### 07-system.sh
- macOS Defaults (Dock, Finder, Keyboard, Trackpad, etc.)
- Fonts
- Mac App Store Apps
- Folder structure
- Scheduled tasks (crontab, LaunchAgents)
- Locale & Timezone
- Login Items
- Hosts file
- Network
- Productivity tools (Raycast, Alfred, Karabiner, Rectangle, BetterTouchTool, 1Password)
- Browser
- AI Tools
- API Tools
- XDG Config
- .env files (age decrypt)
- Database clients
- Package registries

## Script Design

### Shared Preamble (template fragment)
Each script includes:
- `cd "$(dirname "$0")"` — resolve to bundle root
- Architecture check (source vs target)
- Log setup: `~/.machinist/restore-{NN}-{name}.log`
- `log()` and `run_stage()` helper functions
- Stage counter with pass/fail/skip tracking
- Summary on exit with elapsed time

### Orchestrator (install.command)
Minimal — iterates over `0[1-7]-*.sh` files and runs each. No own logic. A group script is only generated if its snapshot sections contain data.

### Each Group Script
- Standalone: can run independently via `bash 03-runtimes.sh`
- Idempotent: each stage checks if already installed
- Exit code: 0 if all stages pass, 1 if any failed
- Logs to own file, summary at end

## Template Changes

| File | Change |
|------|--------|
| `install.command.tmpl` | Rewrite to thin orchestrator |
| `_preamble.sh.tmpl` | New shared fragment with logging, run_stage, arch check |
| `01-foundation.sh.tmpl` | New: includes homebrew, ssh, gpg, git-config templates |
| `02-shell.sh.tmpl` | New: includes shell, terminal, tmux templates |
| `03-runtimes.sh.tmpl` | New: includes node, python, rust, flutter, go, ruby, java, deno, bun, asdf templates |
| `04-editors.sh.tmpl` | New: includes vscode, cursor, xcode, jetbrains, neovim, github-cli templates |
| `05-infrastructure.sh.tmpl` | New: includes docker, k8s, gcp, aws, azure, terraform, firebase, cloudflare, vercel, flyio templates |
| `06-repos.sh.tmpl` | New: includes git-repos template |
| `07-system.sh.tmpl` | New: includes all remaining stage templates |
| `stages/*.sh.tmpl` | Unchanged — existing stage templates reused as-is |

## Code Changes

### `internal/bundler/restorescript.go`
- `GenerateRestoreScript()` → `GenerateRestoreScripts()` returning `map[string]string` (filename → content)
- New helper to render each group template only if its sections have data
- Shared preamble rendered into each group script

### `internal/bundler/dmg.go`
- `PrepareBundleDir()` writes all generated scripts (not just one `install.command`)

### `cmd/machinist/restore.go`
- `--only` and `--skip` flags accept group names (foundation, shell, runtimes, editors, infrastructure, repos, system)
- `--list` shows available groups

### `internal/domain/snapshot.go`
- `StageCount()` remains for backward compat but each group tracks its own count

## What Does NOT Change

- Scanners, snapshot format, domain types
- Stage templates (stages/*.sh.tmpl)
- Config file/dir bundling, encryption
- DMG creation (hdiutil)
- MCP server
