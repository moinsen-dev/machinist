# Machinist README-Parity Implementation Design

> Bring the implementation up to match everything the README promises.

## Scope

- 40+ missing scanners across 7 categories
- CLI command/flag mismatches
- Security pipeline wiring (age encryption is dead code)
- Restore template expansion (13 → 31+ stages)
- Dynamic post-restore checklist
- `applyResult()` deduplication

## Approach

Batch-by-Category: Infrastructure fixes first, then scanners grouped by README category. Each batch is independently testable and committable.

## Batch 0: Infrastructure

### ScanResult Expansion

`scanner.ScanResult` currently has 16 fields for 47 Snapshot sections. Add fields for all missing sections so new scanners have somewhere to put their data.

### applyResult() Deduplication

Two copies exist: `scanner.go:applyResult()` (used by `ScanAll`) and `registry.go:applyResultToSnapshot()` (used by `scan` CLI command). The registry version is missing `Git` and `Docker` fields. Fix: delete `registry.go` version, export `scanner.ApplyResult()` as the single source of truth.

### StageCount() Expansion

Add all new sections to `domain.Snapshot.StageCount()` so the restore script header shows the correct total.

### CLI Fixes

| Fix | Description |
|---|---|
| `snapshot --dry-run` | Add flag: scan + print TOML to stdout without writing file |
| `scan --search-paths` | Add flag for git-repos scanner, pass to `NewGitReposScanner()` |
| `compose` syntax | Keep positional `compose <profile>` (ergonomic). Add `--from-file <manifest.toml>` for loading existing manifests as base. Update README to match. |
| `restore` auto-discover | Make positional arg optional, default to `manifest.toml` in CWD |
| `restore --yes` | Document in README |
| `list` subcommands | Add `list scanners` / `list profiles` sub-subcommands. Keep `list` (no arg) as showing both. Update README. |
| `dmg` command | Document in README Usage section |

### Security Wiring

Wire `internal/security` into the bundler pipeline:
- When `SSHSection` or `GPGSection` is present in snapshot and `Encrypted=true`, encrypt the referenced files using age before bundling into DMG
- `EnvFilesSection` same treatment
- Prompt for age passphrase during `snapshot`/`dmg` command if secret-level files are detected
- Restore script gets a decrypt step before SSH/GPG/env restore stages

### Sensitivity Levels

`ConfigFile.Sensitive` bool already exists. Use it:
- Scanners set `Sensitive=true` on files like `.npmrc`, `.aws/credentials`, `.gitconfig` (contains email)
- Bundler logs a warning for sensitive files being included
- No encryption for sensitive (only secret-level gets encrypted), but user sees what's included

### Dynamic Checklist

Replace static `checklist.md.tmpl` with template logic:
- `{{if .SSH}}` → SSH verification items
- `{{if .GPG}}` → GPG verification items
- `{{if .Network}}` → VPN/DNS items
- `{{if .Browser}}` → Browser sign-in items
- `{{if .Docker}}` → Docker Desktop sign-in
- Base items (TCC permissions, Mac App Store sign-in) always present

### Restore Template Slots

Add `{{if .SectionName}}` blocks in `install.command.tmpl` for every new section. Create stage templates in `templates/stages/` for each scanner category.

## Batch 1: Runtimes (7 scanners)

### go

- Check: `go version`, `go env GOPATH`
- Scan: version, globally installed binaries via `ls $(go env GOPATH)/bin`
- Fills: `GoSection`
- Restore: `go install` for each global package

### java

- Check: `sdk version` (SDKMAN) or `java -version`
- Scan: SDKMAN candidates/versions, default version, JAVA_HOME
- Fills: `JavaSection`
- Restore: install SDKMAN if missing, `sdk install java <version>`

### flutter

- Check: `flutter --version`
- Scan: channel, version, `dart pub global list`
- Fills: `FlutterSection`
- Restore: `flutter channel <channel>`, `dart pub global activate` for each package

### deno

- New domain type: `DenoSection { Version string; GlobalPackages []Package }`
- Check: `deno --version`
- Scan: version, `~/.deno/bin/` contents
- Restore: install via `curl`, `deno install` for each package

### bun

- New domain type: `BunSection { Version string; GlobalPackages []Package }`
- Check: `bun --version`
- Scan: version, `bun pm ls -g`
- Restore: install via `curl`, `bun install -g` for each package

### ruby

- New domain type: `RubySection { Manager string; Versions []string; DefaultVersion string; GlobalGems []Package }`
- Check: `rbenv` or `rvm` or system ruby
- Scan: manager, versions, `gem list`
- Restore: install rbenv/rvm, `rbenv install`, `gem install`

### asdf

- Check: `asdf version` or `mise version`
- Scan: `asdf plugin list`, `asdf list`, `.tool-versions`
- Fills: `AsdfSection`
- Restore: install asdf/mise, add plugins, install versions

## Batch 2: Git & Security (4 scanners)

### git-config

- Scan: `~/.gitconfig`, `~/.gitignore_global`, `git config --global --list` for signing_method, credential_helper, template_dir
- Fills: `GitSection`
- Restore: copy configs, set git config values

### github-cli

- Check: `gh --version`
- Scan: `~/.config/gh/` config dir, `gh extension list`
- Fills: `GitHubCLISection`
- Restore: `gh extension install` for each extension

### ssh

- Check: `~/.ssh/` exists
- Scan: list key files, config, known_hosts
- Set `Encrypted=true`, encrypt key files with age during bundle
- Fills: `SSHSection`
- Restore: decrypt with age, set permissions (600)

### gpg

- Check: `gpg --version`
- Scan: `gpg --list-keys --keyid-format long`, `~/.gnupg/` config files
- Set `Encrypted=true`, encrypt key exports with age
- Fills: `GPGSection`
- Restore: decrypt, `gpg --import`

## Batch 3: Editors (3 scanners + 1 fix)

### jetbrains

- Scan: `~/Library/Application Support/JetBrains/` for IDE directories
- Detect installed IDEs by directory name (IntelliJIdea, PyCharm, WebStorm, GoLand, AndroidStudio, etc.)
- Each IDE: note settings export path
- Fills: `JetBrainsSection`
- Restore: checklist item (settings sync via JetBrains account)

### neovim

- Check: `nvim --version`
- Scan: `~/.config/nvim/` exists, detect plugin manager (lazy.nvim, packer, vim-plug)
- Fills: `NeovimSection`
- Restore: copy config dir, plugin manager auto-installs on first launch

### xcode

- Check: `xcode-select -p`
- Scan: `xcrun simctl list devices -j`, Xcode pref files
- Fills: `XcodeSection`
- Restore: `xcode-select --install`, simulator download commands

### cursor fix

- Add `keybindings.json` detection (like VSCode scanner already does)
- Add `snippets/` directory detection
- No new scanner, just extend existing `CursorScanner`

## Batch 4: Shell & Terminal (3 scanners + 1 fix)

### terminal

- Detect which terminal emulator is in use: iTerm2 (`com.googlecode.iterm2.plist`), Warp (`~/.warp/`), Alacritty (`~/.config/alacritty/`), WezTerm (`~/.config/wezterm/`)
- Capture config files
- Fills: `TerminalSection`
- Restore: copy config files

### tmux

- Check: `tmux -V`
- Scan: `~/.tmux.conf` or `~/.config/tmux/tmux.conf`, parse TPM plugins from `set -g @plugin`
- Fills: `TmuxSection`
- Restore: copy config, clone TPM, `tmux source`

### direnv

- Check: `direnv version`
- Scan: `~/.direnvrc` or `~/.config/direnv/direnvrc`
- Integrate into `ShellSection` as additional config file (no separate domain type — YAGNI)
- Restore: `brew install direnv` (if not via Homebrew already)

### oh-my-zsh custom plugins fix

- Existing ShellConfigScanner: add logic to read `~/.oh-my-zsh/custom/plugins/` directory
- Populate `ShellSection.OhMyZshCustomPlugins` field (already exists)

## Batch 5: Containers & Cloud (11 scanners)

### docker

- Check: `docker --version`
- Scan: `~/.docker/config.json`, detect runtime (Docker Desktop / Colima / OrbStack / Podman)
- `docker images --format '{{.Repository}}:{{.Tag}}'` for frequently used images
- Fills: `DockerSection`
- Restore: install Docker runtime, pull frequently used images

### aws

- Check: `aws --version`
- Scan: `~/.aws/config` exists, `aws configure list-profiles`
- Fills: `AWSSection`
- Restore: copy config file (marked sensitive)

### kubernetes

- Check: `kubectl version --client`
- Scan: `~/.kube/config` exists, `kubectl config get-contexts -o name`
- Fills: `KubernetesSection`
- Restore: copy config file (marked sensitive)

### terraform

- Check: `terraform version`
- Scan: `~/.terraformrc` or `~/.terraform.d/`
- Fills: `TerraformSection`
- Restore: copy config

### vercel

- Check: `vercel --version`
- Scan: `~/.vercel/` directory
- Fills: `VercelSection`
- Restore: copy config dir

### gcp

- New domain type: `GCPSection { ConfigDir string }`
- Check: `gcloud version`
- Scan: `~/.config/gcloud/` exists
- Restore: note in checklist (requires `gcloud auth login`)

### azure

- New domain type: `AzureSection { ConfigDir string }`
- Check: `az version`
- Scan: `~/.azure/` exists
- Restore: note in checklist (requires `az login`)

### flyio

- New domain type: `FlyioSection { ConfigFile string }`
- Check: `fly version`
- Scan: `~/.fly/config.yml`
- Restore: copy config

### firebase

- New domain type: `FirebaseSection { ConfigDir string }`
- Check: `firebase --version`
- Scan: `~/.config/firebase/` or `~/.config/configstore/firebase-tools.json`
- Restore: copy config

### cloudflare

- New domain type: `CloudflareSection { ConfigDir string }`
- Check: `wrangler --version`
- Scan: `~/.config/.wrangler/` or `~/.wrangler/config/`
- Restore: copy config dir

### colima/orbstack/podman

- Not separate scanners — integrated into Docker scanner
- `DockerSection.Runtime` field already captures this
- Detect via: `colima status`, `orbctl status`, `podman --version`

## Batch 6: macOS System (5 scanners + 4 fixes)

### macos-defaults fixes (Trackpad, Mission Control, Spotlight, Menu Bar)

Extend existing `MacOSDefaultsScanner`:
- Trackpad: `defaults read com.apple.AppleMultitouchTrackpad`
- Mission Control: `defaults read com.apple.dock` for hot corners (wvous-* keys)
- Spotlight: `defaults read com.apple.Spotlight orderedItems`
- Menu Bar: `defaults read com.apple.menuextra.clock`, battery percentage

### locale

- Scan: `defaults read NSGlobalDomain AppleLanguages`, `defaults read NSGlobalDomain AppleLocale`, `systemsetup -gettimezone`, `scutil --get ComputerName`, `scutil --get LocalHostName`
- Fills: `LocaleSection`
- Restore: `defaults write`, `systemsetup`, `scutil --set`

### login-items

- Scan: `osascript -e 'tell application "System Events" to get name of every login item'`
- Fills: `LoginItemsSection`
- Restore: checklist item (requires TCC permissions)

### hosts-file

- Scan: read `/etc/hosts`, filter out standard entries (localhost, broadcasthost)
- Fills: `HostsFileSection`
- Restore: append custom entries to `/etc/hosts` (requires sudo)

### network

- Scan: `networksetup -listpreferredwirelessnetworks en0` for Wi-Fi names
- `networksetup -getdnsservers Wi-Fi` for DNS
- `scutil --nc list` for VPN configs
- `tailscale status` if installed
- Fills: `NetworkSection`
- Restore: `networksetup -setdnsservers`, VPN as checklist item

### display

- README mentions "Display" but no domain type exists
- Decision: remove from README (no reliable `defaults write` for display scaling that works across hardware)

## Batch 7: Tools & Config (14 scanners)

### raycast

- Scan: `~/Library/Application Support/com.raycast.macos/` or Raycast export file
- Fills: `RaycastSection`
- Restore: checklist item (import via Raycast app)

### alfred

- New domain type: `AlfredSection { ConfigDir string }`
- Scan: `~/Library/Application Support/Alfred/` or `~/Library/Preferences/com.runningwithcrayons.Alfred-Preferences-3.plist`
- Restore: checklist item (sync via Dropbox/iCloud)

### karabiner

- Scan: `~/.config/karabiner/karabiner.json`
- Fills: `KarabinerSection`
- Restore: copy config file

### rectangle

- Scan: `~/Library/Preferences/com.knewton.Rectangle.plist`
- Fills: `RectangleSection`
- Restore: copy plist

### bettertouchtool

- New domain type: `BetterTouchToolSection { ConfigFile string }`
- Scan: `~/Library/Application Support/BetterTouchTool/` for preset exports
- Restore: checklist item (import via BTT app)

### 1password-cli

- New domain type: `OnePasswordSection { ConfigDir string }`
- Check: `op --version`
- Scan: `~/.config/op/` config
- Restore: `brew install --cask 1password-cli`, checklist for `op signin`

### databases

- Scan: `~/.pgpass` (sensitive), `~/.my.cnf`, TablePlus (`~/Library/Application Support/com.tinyapp.TablePlus/`), DBeaver (`~/.dbeaver4/`)
- Fills: `DatabasesSection`
- Restore: copy config files (mark sensitive)

### registries

- Scan: `~/.npmrc`, `~/.pip/pip.conf`, `~/.cargo/config.toml`, `~/.gemrc`, `~/.cocoapods/config.yaml`, `~/.pub-cache/`
- Fills: `RegistriesSection`
- Restore: copy config files (mark sensitive)

### browser

- Scan: `defaults read com.apple.LaunchServices/com.apple.launchservices.secure LSHandlers` for default browser
- No extension export (not automatable per-browser)
- `ExtensionsChecklist` = static text reminder
- Fills: `BrowserSection`
- Restore: checklist item

### ai-tools

- Scan: `~/.claude/` for Claude Code config, `ollama list` for models
- Fills: `AIToolsSection`
- Restore: copy Claude config, `ollama pull` for each model

### api-tools

- Scan: Postman (`~/Library/Application Support/Postman/`), Insomnia (`~/Library/Application Support/Insomnia/`), `ngrok config check`, `mkcert -CAROOT`
- Fills: `APIToolsSection`
- Restore: copy configs, checklist for API keys

### xdg-config

- Scan: walk `~/.config/` for known tools (bat, lazygit, htop, btop, aerospace, gh, etc.)
- Auto-detect by checking if directory contains config files
- Allow user-specified custom paths
- Fills: `XDGConfigSection`
- Restore: copy config directories

### env-files

- Scan: walk workspace directories for `.env`, `.env.local`, `.env.production` files
- Encrypt with age (secret-level)
- Fills: `EnvFilesSection`
- Restore: decrypt with age, place in original paths

### tailscale

- Integrated into Network scanner
- Check: `tailscale version`
- If installed: add to `NetworkSection` as note
- Restore: `brew install tailscale`, checklist for `tailscale up`

## Testing Strategy

Each scanner gets a `*_test.go` following the existing pattern:
- Mock `CommandRunner` with expected command outputs
- Test happy path (tool installed, data returned)
- Test skip path (tool not installed → nil section)
- Test edge cases (empty output, parse errors)

Integration test: verify `ScanAll` with mocked registry produces a complete Snapshot.

## File Changes Summary

| Area | Files Added | Files Modified |
|---|---|---|
| Domain types | 0 (extend `snapshot.go`, `types.go`) | 2 |
| Scanner infra | 0 | 2 (`scanner.go`, `registry.go`) |
| Batch 1: Runtimes | 14 (7 scanners + 7 tests) | 0 |
| Batch 2: Git & Security | 8 (4 scanners + 4 tests) | 1 (`bundler/dmg.go`) |
| Batch 3: Editors | 4 (2 scanners + 2 tests) | 2 (cursor, xcode) |
| Batch 4: Shell & Terminal | 4 (2 scanners + 2 tests) | 2 (shell, tmux) |
| Batch 5: Cloud | 20 (10 scanners + 10 tests) | 1 (docker) |
| Batch 6: macOS | 8 (4 scanners + 4 tests) | 1 (macos_defaults) |
| Batch 7: Tools | 28 (14 scanners + 14 tests) | 0 |
| CLI fixes | 0 | 5 (snapshot, scan, compose, restore, list) |
| Templates | ~20 stage templates | 3 (install, checklist, README) |
| Total | ~82 new files | ~19 modified files |
