# machinist — Project Architecture Document

> **Mac Developer Environment Snapshot & Restore CLI**
> Scan. Bundle. Restore. Identical dev setup in minutes.

---

## 1. Vision

Ein Entwickler kauft sich einen neuen Mac. Statt tagelang Tools zu installieren, Shell-Configs zu kopieren und Repos zu klonen, führt er **ein DMG** aus – und hat nach 20 Minuten eine identische Arbeitsumgebung.

**machinist** ist ein Rust-CLI-Tool, das:
1. Den aktuellen Mac vollständig scannt (Dev-Tools, System-Settings, Repos, Configs)
2. Einen strukturierten Snapshot als TOML-Manifest erzeugt
3. Diesen Snapshot + alle Config-Files in ein DMG bündelt
4. Auf einem neuen Mac ein enthaltenes Shell-Script ausführt, das alles wiederherstellt

---

## 2. Domain Model

```
┌─────────────────────────────────────────────────┐
│                   Snapshot                       │
│  (Complete picture of a machine's dev env)       │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐   │
│  │  Scanner   │  │  Scanner   │  │  Scanner   │  │
│  │ (Homebrew) │  │  (Shell)   │  │  (Git)     │  │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘   │
│        │              │              │           │
│        ▼              ▼              ▼           │
│  ┌─────────────────────────────────────────┐    │
│  │            Manifest (TOML)              │    │
│  │  Packages, Configs, Repos, Defaults...  │    │
│  └─────────────────────────────────────────┘    │
│                      │                           │
│                      ▼                           │
│  ┌─────────────────────────────────────────┐    │
│  │              Bundler                     │    │
│  │   Manifest + Files → DMG                │    │
│  └─────────────────────────────────────────┘    │
└─────────────────────────────────────────────────┘
```

### Core Entities

| Entity | Beschreibung |
|---|---|
| **Snapshot** | Gesamtbild einer Maschine — enthält alle ScanResults |
| **Scanner** | Trait — jeder Scanner ist für eine Domain zuständig |
| **ScanResult** | Was ein Scanner gefunden hat (Packages, Configs, etc.) |
| **Manifest** | Serialisierbare TOML-Repräsentation des Snapshots |
| **Bundler** | Nimmt Manifest + gesammelte Files, erzeugt DMG |
| **RestoreScript** | Generiertes Shell-Script, das den Snapshot wiederherstellt |

### Value Objects

```rust
Package       { name, version, source: PackageSource }
Repository    { path, remote_url, branch, shallow: bool }
ConfigFile    { source_path, category, content_hash }
MacDefault    { domain, key, value, value_type }
InstalledApp  { name, source: AppSource, bundle_id: Option }
Font          { name, file_path }
CronJob       { schedule, command, source: CronSource }
EnvFile       { path, encrypted: bool }
```

---

## 3. Scanner-Module

Jeder Scanner implementiert den `Scanner`-Trait und ist unabhängig testbar.
Scanner sind in **Kategorien** organisiert, jede Kategorie kann mehrere Sub-Scanner haben.

### 3.1 Package Managers & Runtimes

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Homebrew** | Taps, Formulae, Casks, `brew services` (laufende Hintergrund-Dienste wie postgres, redis) | `brew tap`, `brew install`, `brew install --cask`, `brew services start` |
| **npm/pnpm/yarn** | Global packages, nvm/fnm Node-Versionen, default version | `nvm install`, `npm i -g` |
| **Python** | pyenv/uv versions, pip global packages, virtualenvs (Pfade + requirements.txt), conda envs | `pyenv install`, `pip install`, `uv sync` |
| **Rust** | rustup toolchains + components, cargo-installed binaries | `rustup toolchain install`, `cargo install` |
| **Ruby** | rbenv/rvm versions, global gems | `rbenv install`, `gem install` |
| **Go** | Go version via homebrew/goenv, go install'd binaries | `go install` |
| **Java/Kotlin** | SDKMAN config (~/.sdkman/), JDK versions, JAVA_HOME | `sdk install java`, `sdk default` |
| **Flutter/Dart** | Flutter SDK channel + version, dart pub global packages | `flutter channel`, `dart pub global activate` |
| **Deno** | Deno version, deno install'd tools | `deno install` |
| **Bun** | Bun version, global installs | `bun install -g` |
| **asdf/mise** | ~/.tool-versions, installed plugins + versions (universeller Version Manager) | `asdf plugin add`, `asdf install` |

### 3.2 Shell & Terminal

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Shell Config** | .zshrc, .zshenv, .zprofile, .zlogout, .bashrc, .bash_profile, .inputrc, .hushlogin, custom function files (~/.functions, ~/.aliases) | Dateien kopieren |
| **Shell Framework** | oh-my-zsh (custom plugins, themes), prezto, etc. | git clone + Dateien kopieren |
| **Prompt** | starship.toml, powerlevel10k config, pure prompt | Dateien kopieren |
| **Terminal Emulator** | iTerm2 Profile + Color Schemes (JSON export), Terminal.app Profile, Warp config, Alacritty config (~/.config/alacritty/) | Dateien kopieren, iTerm2 `defaults import` |
| **tmux** | ~/.tmux.conf, TPM plugins liste, custom scripts | Dateien kopieren, `git clone tpm`, `tpm install` |
| **direnv** | ~/.config/direnv/direnvrc, global .envrc patterns | Dateien kopieren |

### 3.3 Git & Version Control

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Git Config** | ~/.gitconfig, ~/.gitignore_global, ~/.gitattributes | Dateien kopieren |
| **Git Templates** | ~/.git-templates/ (hooks, commit templates) | Verzeichnis kopieren |
| **Git Signing** | GPG keys für Commit-Signing, SSH signing config | `gpg --import`, config kopieren |
| **Git Credential Helper** | osxkeychain config, credential cache settings | Config kopieren |
| **GitHub CLI** | ~/.config/gh/ (config.yml, hosts.yml) — AUTH TOKENS ENCRYPTED | `gh auth login` Anleitung + config kopieren |
| **Git Repos** | Alle Git-Repos (konfigurierbare Suchpfade), Remote-URL, aktueller Branch, Submodule-Info | `git clone [--recurse-submodules]`, optional shallow |
| **Pre-commit** | ~/.cache/pre-commit/ global hooks config | Config kopieren |

### 3.4 Editors & IDEs

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **VSCode** | Extensions, settings.json, keybindings.json, snippets/, profiles | `code --install-extension`, Dateien kopieren |
| **Cursor** | Gleich wie VSCode aber unter ~/.cursor/, Cursor-spezifische AI-Settings | Dateien kopieren |
| **JetBrains** | Installierte IDEs (IntelliJ, PyCharm, WebStorm, Android Studio), Plugin-Listen, exportierte Settings (.zip), Keymaps, Code Styles, Live Templates | Settings-Import, Plugin-Install |
| **Neovim** | ~/.config/nvim/ komplett (init.lua/vim, lazy.nvim/packer plugins) | Verzeichnis kopieren, Plugin-Manager sync |
| **Vim** | ~/.vimrc, ~/.vim/ | Dateien kopieren |
| **Xcode** | Custom key bindings, code snippets, themes, installed simulators list | Dateien kopieren, `xcrun simctl` |

### 3.5 Container & Virtualization

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Docker** | ~/.docker/config.json (registries, auth — ENCRYPTED), Docker Desktop settings, häufig genutzte Images (liste) | Config kopieren, `docker pull` für gelistete Images |
| **Colima** | ~/.colima/ config | Config kopieren, `colima start` |
| **OrbStack** | Settings + config | Config kopieren |
| **Podman** | ~/.config/containers/ | Config kopieren |

### 3.6 Cloud & DevOps

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **AWS CLI** | ~/.aws/config (Profile, Regions — OHNE credentials), ~/.aws/credentials wird ENCRYPTED oder nur Profil-Struktur | Config kopieren, Credentials separat encrypted |
| **GCP** | ~/.config/gcloud/ (properties, configurations — OHNE tokens) | Config kopieren, `gcloud auth login` Anleitung |
| **Azure** | ~/.azure/ config | Config kopieren |
| **Kubernetes** | ~/.kube/config (Cluster-Definitionen — OHNE Tokens), Helm repos | Config-Struktur kopieren, `helm repo add` |
| **Terraform** | ~/.terraformrc, provider cache config | Config kopieren |
| **Vercel** | ~/.config/vercel/ | Config kopieren |
| **Fly.io** | ~/.fly/ config | Config kopieren |
| **Firebase** | ~/.config/firebase/ | Config kopieren |
| **Cloudflare** | Wrangler config | Config kopieren |

### 3.7 Database Clients

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **PostgreSQL** | ~/.pgpass (ENCRYPTED), ~/.psqlrc | Dateien kopieren |
| **MySQL** | ~/.my.cnf (ENCRYPTED) | Dateien kopieren |
| **TablePlus** | Connection-Export (ohne Passwörter) | Import |
| **DBeaver** | ~/.dbeaver/ connections (ohne Passwörter) | Config kopieren |
| **Redis** | redis-cli config | Config kopieren |

### 3.8 Package Registry Auth & Config

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **npm** | ~/.npmrc (Registry-URLs, Scopes — Tokens ENCRYPTED) | Dateien kopieren |
| **pip** | ~/.pip/pip.conf, ~/.config/pip/ | Dateien kopieren |
| **Cargo** | ~/.cargo/config.toml (custom registries, build settings) | Dateien kopieren |
| **Gem** | ~/.gemrc | Dateien kopieren |
| **CocoaPods** | Podfile-Templates, CDN config | Config kopieren |
| **Pub (Dart)** | ~/.pub-cache/ config | Config kopieren |

### 3.9 macOS System

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Dock** | Size, Position, Auto-Hide, angeheftete Apps (Bundle-IDs) | `defaults write com.apple.dock` |
| **Finder** | Show hidden, Path bar, Status bar, Default view, Sidebar items | `defaults write com.apple.finder` |
| **Keyboard** | Key repeat rate, Initial repeat delay, custom shortcuts, Input sources | `defaults write NSGlobalDomain`, `defaults write com.apple.symbolichotkeys` |
| **Trackpad** | Tap-to-click, Tracking speed, Gestures | `defaults write` |
| **Mission Control** | Hot Corners, Spaces config, "Displays have separate Spaces" | `defaults write com.apple.dock` |
| **Accessibility** | Reduce motion, Font size, Cursor size | `defaults write` |
| **Login Items** | Apps die bei Login starten | `osascript` oder LaunchAgent |
| **Hosts File** | Custom /etc/hosts Einträge (z.B. für lokale Dev-Domains) | Append zu /etc/hosts (sudo) |
| **Homebrew Services** | Aktuell laufende Services (postgres, redis, nginx, etc.) | `brew services start` |
| **Spotlight** | Exclusion-Pfade (z.B. node_modules) | `defaults write com.apple.Spotlight` |
| **Screenshots** | Default-Pfad, Format (PNG/JPG), Schatten an/aus | `defaults write com.apple.screencapture` |
| **Locale** | Sprache, Region, Zeitzone, Kalender-Format | `defaults write NSGlobalDomain`, `systemsetup -settimezone` |
| **Sharing** | ComputerName, LocalHostName | `scutil --set ComputerName`, `scutil --set LocalHostName` |
| **Display** | Scaled Resolution, Night Shift Schedule | `defaults write` (teilweise), Hinweis für manuelle Settings |
| **Menu Bar** | Clock-Format, Batterie-%, Control Center Items | `defaults write com.apple.menuextra.*` |

### 3.10 Productivity & Power User Tools

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Raycast** | Extensions, Snippets, Konfig-Export (Raycast hat JSON export) | Import via Raycast |
| **Alfred** | Workflows, Preferences, Snippets, Custom web searches | Sync-Folder kopieren |
| **Karabiner-Elements** | ~/.config/karabiner/ (komplexe Key-Remappings!) | Config kopieren |
| **Rectangle/Magnet** | Window-Management Shortcuts + Config | `defaults write` oder Config kopieren |
| **BetterTouchTool** | Presets export | Config kopieren |
| **1Password CLI** | CLI config (OHNE Vault-Daten) | Config kopieren, `op signin` Anleitung |

### 3.11 Security & Credentials

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **SSH** | SSH keys (ENCRYPTED mit age), ~/.ssh/config, known_hosts | Encrypted copy, Dateien kopieren |
| **GPG** | GPG keys export (ENCRYPTED), gpg-agent.conf | `gpg --import`, Config kopieren |
| **age** | age keys falls vorhanden | Encrypted copy |
| **Keychain Notes** | KEINE automatische Erfassung — nur Hinweis an User | Manuell |

### 3.12 XDG Config Catch-All Scanner

Für alles unter `~/.config/` was nicht von einem spezifischen Scanner erfasst wird:

| Was | Beispiele |
|---|---|
| **Known Patterns** | bat/, lazygit/, htop/, btop/, lsd/, eza/, fd/, ripgreprc, wezterm/, aerospace/ |
| **Detection** | Liest ~/.config/, matched gegen Registry bekannter Tool-Configs |
| **Erweiterbar** | User kann in machinist.toml eigene Pfade hinzufügen |
| **Restore** | Verzeichnisse/Dateien 1:1 kopieren |

```toml
# User kann in machinist.toml eigene Pfade ergänzen
[xdg_config]
auto_detected = ["bat", "lazygit", "htop", "btop", "aerospace"]
custom_paths = [
    "~/.config/my-custom-tool/",
    "~/.my-obscure-rc",
]
```

### 3.13 Workspace & Projects

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Folder Structure** | Workspace-Verzeichnisstruktur (konfigurierbare Tiefe + Pfade) | `mkdir -p` |
| **Git Repos** | Alle Git-Repos, Remote-URL, Branch, Submodule-Status | `git clone`, optional `--depth 1` |
| **Environment Files** | .env, .env.local etc. in Projekten (ENCRYPTED, optional, explicit opt-in) | Encrypted copy |
| **Workspace Files** | .editorconfig, .prettierrc, .eslintrc in Home-Dir (globale Configs) | Dateien kopieren |

### 3.14 Fonts

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **User Fonts** | ~/Library/Fonts/ (alle manuell installierten Fonts) | Dateien kopieren nach ~/Library/Fonts/ |
| **Nerd Fonts** | Erkennung ob via Homebrew Cask installiert (wird dann dort behandelt) | via Homebrew |

### 3.15 Scheduled Tasks

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Crontab** | User crontab (`crontab -l`) | `crontab` import |
| **LaunchAgents** | ~/Library/LaunchAgents/*.plist (User-Level) | plist kopieren, `launchctl load` |
| **LaunchDaemons** | KEINE (System-Level, zu riskant) | — |

### 3.16 Network & VPN

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Wi-Fi** | Gespeicherte Netzwerke + Prioritäten (`networksetup -listpreferredwirelessnetworks`) | `networksetup` (Passwörter müssen manuell eingegeben werden) |
| **VPN** | VPN-Konfigurationen (IKEv2, WireGuard, OpenVPN Profiles) | Config kopieren, `networksetup -importpwf` oder WireGuard-Config |
| **DNS** | Custom DNS-Server pro Interface, lokale resolver configs (dnsmasq) | `networksetup -setdnsservers` |
| **Proxy** | HTTP/SOCKS Proxy Settings, PAC files | `networksetup -setwebproxy` |
| **Tailscale** | Installiert via Homebrew Cask (dort erfasst), Config-Hinweis | Cask-Install + `tailscale up` Anleitung |

### 3.17 Browser

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Default Browser** | Welcher Browser ist Default | `defaultbrowser` CLI oder `open -a` Hint |
| **Browser Extensions** | Liste installierter Extensions (Chrome, Firefox, Arc, Safari) | Extensions-Liste als Checkliste (kein Auto-Install möglich) |
| **Dev Extensions** | React DevTools, Redux DevTools, Vue DevTools etc. | Checkliste generieren |

### 3.18 API & Dev Tools

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Postman/Insomnia** | Collection-Exports, Environments (ENCRYPTED) | Import-Anleitung + Files |
| **ngrok** | `~/.ngrok2/ngrok.yml` (Auth-Token ENCRYPTED) | Config kopieren |
| **mkcert** | Lokale CA für HTTPS Dev | `mkcert -install` im Restore-Script |

### 3.19 AI Developer Tools

| Scanner | Was wird erfasst | Restore-Strategie |
|---|---|---|
| **Claude Code** | `~/.claude/` (Settings, Memory, CLAUDE.md) | Config kopieren |
| **Ollama** | Installierte Modelle-Liste (nicht die Modelle selbst!) | `ollama pull` pro Modell |
| **GitHub Copilot** | Via VSCode/Editor-Scanner abgedeckt | — |

---

## 4. CLI Interface

```bash
# Vollständiger Snapshot → DMG
machinist snapshot
machinist snapshot --output ~/Desktop/machinist.dmg
machinist snapshot --interactive          # Fragt pro Scanner nach Include/Exclude
machinist snapshot --dry-run              # Zeigt was erfasst würde, ohne zu exportieren

# Einzelne Scanner ausführen (Debug/Inspect)
machinist scan homebrew
machinist scan shell
machinist scan git-repos --search-paths ~/Code,~/Projects

# Restore (auf neuem Mac, vom DMG aus)
machinist restore                         # Liest machinist.toml aus aktuellem Verzeichnis
machinist restore --skip homebrew,fonts   # Bestimmte Scanner überspringen
machinist restore --dry-run               # Zeigt was installiert würde

# Info
machinist list-scanners                   # Zeigt alle verfügbaren Scanner
machinist version
```

---

## 5. Manifest Format (machinist.toml)

Das Manifest ist modular aufgebaut — jede Scanner-Kategorie hat ihre eigene TOML-Section.
Nur Sections die beim Scan Ergebnisse liefern werden ins Manifest geschrieben.

```toml
[meta]
created_at = "2026-02-18T14:30:00Z"
source_hostname = "Moinsen-MacBook-Pro"
source_os_version = "15.3"
source_arch = "arm64"                     # wichtig für Intel vs Apple Silicon
machinist_version = "0.1.0"
scan_duration_secs = 12

# ─── PACKAGE MANAGERS ───────────────────────────────────────

[homebrew]
taps = ["homebrew/core", "homebrew/cask", "dart-lang/dart"]
formulae = [
    { name = "git", version = "2.43.0" },
    { name = "ripgrep", version = "14.1.0" },
    { name = "fd", version = "9.0.0" },
    { name = "postgresql@16", version = "16.1" },
]
casks = [
    { name = "visual-studio-code" },
    { name = "iterm2" },
    { name = "docker" },
    { name = "raycast" },
    { name = "karabiner-elements" },
]
services = [
    { name = "postgresql@16", status = "started" },
    { name = "redis", status = "started" },
]

[node]
manager = "fnm"                           # "nvm" | "fnm" | "volta" | "asdf"
versions = ["20.11.0", "18.19.0"]
default_version = "20.11.0"
global_packages = [
    { name = "typescript", version = "5.3.3" },
    { name = "pnpm", version = "8.14.0" },
    { name = "vercel", version = "33.0.0" },
]

[python]
manager = "uv"                            # "pyenv" | "uv" | "conda" | "asdf"
versions = ["3.12.1", "3.11.7"]
default_version = "3.12.1"
global_packages = [
    { name = "black", version = "24.1.0" },
    { name = "ruff", version = "0.1.14" },
    { name = "httpie", version = "3.2.2" },
]

[rust]
toolchains = ["stable", "nightly"]
default_toolchain = "stable"
components = ["clippy", "rustfmt", "rust-analyzer"]
cargo_packages = [
    { name = "cargo-watch", version = "8.5.2" },
    { name = "cargo-edit", version = "0.12.2" },
    { name = "cargo-nextest", version = "0.9.67" },
]

[java]
manager = "sdkman"                        # "sdkman" | "asdf" | "manual"
versions = ["21.0.1-tem", "17.0.9-tem"]
default_version = "21.0.1-tem"
java_home = "~/.sdkman/candidates/java/current"

[flutter]
channel = "stable"
version = "3.19.0"
dart_global_packages = ["melos", "very_good_cli"]

[go]
version = "1.22.0"
global_packages = [
    { name = "golang.org/x/tools/gopls", version = "latest" },
]

[asdf]
plugins = [
    { name = "elixir", versions = ["1.16.0"] },
    { name = "erlang", versions = ["26.2.1"] },
]
tool_versions_file = "configs/asdf/.tool-versions"

# ─── SHELL & TERMINAL ──────────────────────────────────────

[shell]
default_shell = "/bin/zsh"
framework = "oh-my-zsh"
prompt = "starship"
config_files = [
    { source = "~/.zshrc", bundle_path = "configs/shell/.zshrc" },
    { source = "~/.zshenv", bundle_path = "configs/shell/.zshenv" },
    { source = "~/.zprofile", bundle_path = "configs/shell/.zprofile" },
    { source = "~/.aliases", bundle_path = "configs/shell/.aliases" },
    { source = "~/.functions", bundle_path = "configs/shell/.functions" },
    { source = "~/.hushlogin", bundle_path = "configs/shell/.hushlogin" },
    { source = "~/.config/starship.toml", bundle_path = "configs/shell/starship.toml" },
]
oh_my_zsh_custom_plugins = ["zsh-autosuggestions", "zsh-syntax-highlighting"]

[terminal]
app = "iterm2"                            # "iterm2" | "terminal" | "warp" | "alacritty" | "wezterm"
config_files = [
    { source = "~/Library/Preferences/com.googlecode.iterm2.plist", bundle_path = "configs/terminal/iterm2.plist" },
]

[tmux]
config_files = [
    { source = "~/.tmux.conf", bundle_path = "configs/tmux/.tmux.conf" },
]
tpm_plugins = ["tmux-plugins/tpm", "tmux-plugins/tmux-sensible", "tmux-plugins/tmux-resurrect"]

# ─── GIT & VERSION CONTROL ─────────────────────────────────

[git]
config_files = [
    { source = "~/.gitconfig", bundle_path = "configs/git/.gitconfig" },
    { source = "~/.gitignore_global", bundle_path = "configs/git/.gitignore_global" },
]
signing_method = "ssh"                    # "gpg" | "ssh" | "none"
template_dir = "configs/git/templates/"
credential_helper = "osxkeychain"

[github_cli]
config_dir = "configs/gh/"                # ~/.config/gh/ (config.yml — OHNE auth tokens)
extensions = ["gh-dash", "gh-copilot"]

[git_repos]
search_paths = ["~/Code", "~/Projects"]
repositories = [
    { path = "~/Code/machinist", remote = "git@github.com:moinsen/machinist.git", branch = "main", shallow = false },
    { path = "~/Code/my-app", remote = "git@github.com:moinsen/my-app.git", branch = "main", shallow = false },
    { path = "~/Code/huge-monorepo", remote = "git@github.com:company/monorepo.git", branch = "main", shallow = true },
]

# ─── EDITORS & IDEs ─────────────────────────────────────────

[vscode]
extensions = [
    "rust-analyzer",
    "dart-code.flutter",
    "bradlc.vscode-tailwindcss",
    "github.copilot",
]
config_files = [
    { source = "settings.json", bundle_path = "configs/vscode/settings.json" },
    { source = "keybindings.json", bundle_path = "configs/vscode/keybindings.json" },
]
snippets_dir = "configs/vscode/snippets/"

[cursor]
extensions = ["cursor-specific-ext"]
config_files = [
    { source = "settings.json", bundle_path = "configs/cursor/settings.json" },
]

[neovim]
config_dir = "configs/nvim/"              # komplettes ~/.config/nvim/
plugin_manager = "lazy.nvim"

[jetbrains]
ides = [
    { name = "IntelliJ IDEA", settings_export = "configs/jetbrains/intellij-settings.zip" },
    { name = "Android Studio", settings_export = "configs/jetbrains/android-studio-settings.zip" },
]

[xcode]
simulators = ["iPhone 15 Pro", "iPad Pro (12.9-inch)"]
config_files = [
    { source = "~/Library/Developer/Xcode/UserData/KeyBindings/", bundle_path = "configs/xcode/keybindings/" },
]

# ─── CONTAINERS ─────────────────────────────────────────────

[docker]
config_file = "configs/docker/config.json"    # Registries, OHNE Auth (encrypted separat)
frequently_used_images = [
    "postgres:16-alpine",
    "redis:7-alpine",
    "node:20-alpine",
]
runtime = "docker-desktop"                # "docker-desktop" | "colima" | "orbstack"

# ─── CLOUD & DEVOPS ────────────────────────────────────────

[aws]
config_file = "configs/aws/config"        # Profile + Regions, OHNE credentials
profiles = ["default", "staging", "production"]
# Credentials werden separat encrypted oder weggelassen

[kubernetes]
config_file = "configs/kube/config"       # Cluster-Definitionen, Contexts (OHNE Tokens)
contexts = ["docker-desktop", "staging-cluster", "prod-cluster"]

[terraform]
config_file = "configs/terraform/.terraformrc"

[vercel]
config_dir = "configs/vercel/"

# ─── DATABASE CLIENTS ──────────────────────────────────────

[databases]
config_files = [
    { source = "~/.psqlrc", bundle_path = "configs/db/.psqlrc" },
    { source = "~/.pgpass", bundle_path = "configs/db/.pgpass.age", encrypted = true },
]

# ─── PACKAGE REGISTRY CONFIG ──────────────────────────────

[registries]
config_files = [
    { source = "~/.npmrc", bundle_path = "configs/registries/.npmrc", sensitive = true },
    { source = "~/.cargo/config.toml", bundle_path = "configs/registries/cargo-config.toml" },
    { source = "~/.pip/pip.conf", bundle_path = "configs/registries/pip.conf" },
    { source = "~/.gemrc", bundle_path = "configs/registries/.gemrc" },
]

# ─── macOS SYSTEM ──────────────────────────────────────────

[macos_defaults]
dock = { autohide = true, tilesize = 48, orientation = "bottom", minimize_effect = "scale" }
finder = { show_hidden = true, show_path_bar = true, show_status_bar = true, default_view = "list" }
keyboard = { key_repeat = 2, initial_key_repeat = 15 }
trackpad = { tap_to_click = true, tracking_speed = 3.0 }
mission_control = { hot_corners = { top_left = "mission-control", bottom_right = "desktop" } }
spotlight = { excluded_paths = ["~/Code/*/node_modules", "~/Code/*/.build"] }
screenshots = { path = "~/Desktop/Screenshots", format = "png", disable_shadow = true }
menu_bar = { clock_format = "HH:mm", show_battery_percentage = true }

[locale]
language = "de-DE"
region = "DE"
timezone = "Europe/Berlin"
computer_name = "Moinsen-MacBook-Pro"
local_hostname = "Moinsen-MacBook-Pro"

[login_items]
apps = ["Docker", "Raycast", "Rectangle", "Bartender"]

[hosts_file]
custom_entries = [
    { ip = "127.0.0.1", hostnames = ["local.myapp.dev", "api.local.myapp.dev"] },
]

[apps]
app_store = [
    { name = "Xcode", id = 497799835 },
    { name = "Magnet", id = 441258766 },
]

# ─── PRODUCTIVITY TOOLS ────────────────────────────────────

[raycast]
export_file = "configs/raycast/raycast-export.json"
# Raycast hat einen eingebauten Export/Import Mechanismus

[karabiner]
config_dir = "configs/karabiner/"         # ~/.config/karabiner/

[rectangle]
config_file = "configs/rectangle/RectangleConfig.json"

# ─── SECURITY ──────────────────────────────────────────────

[ssh]
encrypted = true
config_file = "configs/ssh/config"
keys = ["configs/ssh/id_ed25519.age", "configs/ssh/id_ed25519.pub"]
known_hosts = "configs/ssh/known_hosts"

[gpg]
encrypted = true
keys = ["configs/gpg/private-keys.age"]
config_files = [
    { source = "~/.gnupg/gpg-agent.conf", bundle_path = "configs/gpg/gpg-agent.conf" },
]

# ─── XDG CONFIG CATCH-ALL ──────────────────────────────────

[xdg_config]
auto_detected = ["bat", "lazygit", "htop", "btop", "aerospace", "fd", "ripgrep"]
custom_paths = []
config_dir = "configs/xdg/"

# ─── WORKSPACE ─────────────────────────────────────────────

[folders]
structure = [
    "~/Code",
    "~/Code/personal",
    "~/Code/work",
    "~/Code/oss",
    "~/Projects",
    "~/Scripts",
]

[fonts]
custom_fonts = [
    { name = "FiraCode Nerd Font", bundle_path = "configs/fonts/FiraCode-NF.ttf" },
    { name = "JetBrains Mono", bundle_path = "configs/fonts/JetBrainsMono.ttf" },
]

[env_files]
encrypted = true
files = [
    { source = "~/Code/my-app/.env", bundle_path = "configs/env/my-app.env.age" },
]

# ─── NETWORK & VPN ────────────────────────────────────────

[network]
preferred_wifi = ["HomeNetwork", "Office-5G"]
dns = { interface = "Wi-Fi", servers = ["1.1.1.1", "8.8.8.8"] }
vpn_configs = [
    { source = "~/.config/wireguard/wg0.conf", bundle_path = "configs/network/wg0.conf", encrypted = true },
]

# ─── BROWSER ──────────────────────────────────────────────

[browser]
default = "arc"                               # Browser Bundle-ID oder Name
extensions_checklist = "configs/browser/extensions.md"

# ─── AI DEVELOPER TOOLS ──────────────────────────────────

[ai_tools]
claude_code_config = "configs/ai/claude/"     # ~/.claude/ (Settings, Memory)
ollama_models = ["llama3:8b", "codellama:13b"]

# ─── API & DEV TOOLS ─────────────────────────────────────

[api_tools]
config_files = [
    { source = "~/.ngrok2/ngrok.yml", bundle_path = "configs/api/ngrok.yml", encrypted = true },
]
mkcert = true                                 # `mkcert -install` im Restore ausführen

# ─── SCHEDULED TASKS ──────────────────────────────────────

[crontab]
entries = "configs/cron/crontab.txt"

[launchagents]
plists = [
    { source = "~/Library/LaunchAgents/com.myapp.sync.plist", bundle_path = "configs/launchagents/com.myapp.sync.plist" },
]
```

---

## 6. DMG-Struktur

```
machinist.dmg/
├── install.command            # Doppelklickbar im Finder (macOS .command extension)
├── machinist.toml             # Das Manifest
├── configs/                   # Gesammelte Config-Files
│   ├── shell/
│   │   ├── .zshrc
│   │   ├── .zprofile
│   │   └── starship.toml
│   ├── vscode/
│   │   ├── settings.json
│   │   └── keybindings.json
│   ├── git/
│   │   ├── .gitconfig
│   │   └── .gitignore_global
│   ├── ssh/                   # Encrypted mit age
│   │   ├── config
│   │   ├── id_ed25519.age
│   │   └── id_ed25519.pub
│   ├── fonts/
│   │   └── *.ttf / *.otf
│   ├── env/                   # Encrypted mit age
│   │   └── *.env.age
│   ├── network/               # VPN configs (encrypted)
│   │   └── wg0.conf
│   ├── ai/                    # AI Tool configs
│   │   └── claude/
│   ├── api/                   # API tool configs
│   │   └── ngrok.yml
│   ├── browser/
│   │   └── extensions.md      # Generierte Checkliste
│   └── launchagents/
│       └── *.plist
├── post-restore-checklist.md  # Manuelle Schritte nach Restore
└── README.md                  # Anleitung für den User
```

---

## 7. Restore-Script Architektur

Das `install.command` Shell-Script wird **von machinist generiert** und ist idempotent.

### Restore-Reihenfolge (Stages)

Die Reihenfolge ist kritisch — Dependencies müssen vor Dependents kommen.

```
 Stage  1: Xcode Command Line Tools + Rosetta 2  (Voraussetzung für alles; Rosetta nur auf Apple Silicon)
 Stage  2: SSH Keys entschlüsseln            (FRÜH — wird für Git Clone gebraucht)
 Stage  3: GPG Keys importieren              (für Git Signing)
 Stage  4: Homebrew installieren
 Stage  5: Homebrew Taps + Formulae + Casks  (installiert auch Docker, iTerm2, Raycast etc.)
 Stage  6: Homebrew Services starten         (postgres, redis etc.)
 Stage  7: Shell-Konfiguration               (.zshrc, oh-my-zsh, starship, tmux)
 Stage  8: Language Runtimes                  (nvm/fnm, pyenv/uv, rustup, sdkman, asdf)
 Stage  9: Language Packages                  (npm globals, pip, cargo install, dart globals)
 Stage 10: Folder-Struktur erzeugen
 Stage 11: Git Repos klonen                   (braucht SSH Keys + Folder-Struktur)
 Stage 12: Git Config + Templates
 Stage 13: Editor/IDE Setup                   (VSCode Extensions, Cursor, Neovim, JetBrains)
 Stage 14: AI Developer Tools                  (Claude Code Config, Ollama Models pullen)
 Stage 15: XDG Configs kopieren               (bat, lazygit, htop etc.)
 Stage 16: Cloud CLI Configs                   (AWS, GCP, kubectl)
 Stage 17: Database Client Configs
 Stage 18: Package Registry Configs            (.npmrc, pip.conf, cargo config)
 Stage 19: Productivity Tools                  (Raycast import, Karabiner, Rectangle)
 Stage 20: Fonts installieren
 Stage 21: macOS Defaults setzen               (Dock, Finder, Keyboard, Trackpad, Screenshots, Menu Bar)
 Stage 22: Locale & System Identity            (Sprache, Region, Timezone, ComputerName)
 Stage 23: Login Items konfigurieren
 Stage 24: Hosts File Einträge                 (braucht sudo)
 Stage 25: Network & VPN                       (DNS, Proxy, VPN-Configs, Wi-Fi Prioritäten)
 Stage 26: App Store Apps (via mas)
 Stage 27: LaunchAgents / Cron
 Stage 28: Environment Files entschlüsseln     (optional)
 Stage 29: Docker Images pullen                (optional, kann lang dauern)
 Stage 30: Verify & Summary                    (Prüft was erfolgreich war, was fehlschlug)
 Stage 31: Post-Restore Checkliste ausgeben     (manuelle Schritte die nicht automatisierbar sind)
```

Jede Stage:
- Gibt Fortschritt aus (Stage X/31: "Installing Homebrew packages...")
- Ist überspringbar (`--skip stage-name`)
- Ist idempotent (prüft ob schon installiert/vorhanden)
- Loggt in `~/.machinist/restore.log`
- Hat ein Timeout (konfigurierbar, default 10min pro Stage)
- Bei Fehler: loggen + weiter zur nächsten Stage (kein Abbruch)

### 7.1 Post-Restore Checkliste

Stage 31 generiert eine Checkliste mit Dingen die **nicht automatisierbar** sind. Diese wird sowohl in der Terminal-Ausgabe als auch als `~/Desktop/machinist-post-restore.md` gespeichert.

| Kategorie | Was manuell erledigt werden muss |
|---|---|
| **TCC Permissions** | Accessibility, Full Disk Access, Screen Recording pro App freischalten (Karabiner, Terminal, Raycast etc.) |
| **Gatekeeper** | `xattr -d com.apple.quarantine` für CLI-Tools die beim ersten Start geblockt werden |
| **Browser Extensions** | Manuelle Installation aus generierter Extensions-Liste |
| **Browser Profiles** | Work/Personal Profile einrichten, Bookmarks importieren |
| **iCloud / Cloud Sync** | iCloud Drive, Dropbox, Google Drive manuell einrichten |
| **Communication Tools** | Slack Workspaces beitreten, Zoom/Teams/Discord Login |
| **VPN Passwörter** | VPN-Credentials manuell eingeben |
| **2FA / Auth Apps** | Authenticator-App einrichten, Recovery Codes |
| **Printer** | Drucker neu einrichten |
| **Bluetooth** | Geräte neu koppeln (Tastatur, Maus, Kopfhörer) |

---

## 8. Security

| Thema | Lösung |
|---|---|
| SSH Keys | Verschlüsselt mit [age](https://github.com/FiloSottile/age) — Passphrase bei Snapshot setzen, bei Restore eingeben |
| .env Files | Gleiche age-Verschlüsselung |
| Sensitive Defaults | Scanner fragt explizit nach ("SSH Keys einschließen? [y/N]") |
| DMG-Passwort | Optional: DMG selbst mit Passwort verschlüsseln via `hdiutil` |

---

## 9. Architektur-Entscheidungen

| Entscheidung | Begründung |
|---|---|
| **Rust für Scanner + Bundler** | Schnell, typsicher, gutes macOS Ecosystem, single binary |
| **Shell-Script für Restore** | Muss auf Vanilla-Mac laufen ohne Rust/Cargo |
| **TOML als Manifest** | Human-readable, human-editable, Rust-native (serde) |
| **age für Encryption** | Modern, simpel, kein GPG-Overhead |
| **Modulare Scanner** | Jeder Scanner unabhängig, leicht erweiterbar |
| **hdiutil für DMG** | macOS-native, kein extra Dependency |

---

## 10. Rust Project Structure

Scanner sind in Submodule nach Kategorie gruppiert — hält die Codebase übersichtlich.

```
machinist/
├── Cargo.toml
├── src/
│   ├── main.rs                         # CLI entry point (clap)
│   ├── lib.rs                          # Public API
│   │
│   ├── cli/
│   │   ├── mod.rs
│   │   ├── snapshot.rs                 # snapshot subcommand
│   │   ├── scan.rs                     # scan subcommand
│   │   └── restore.rs                  # restore subcommand (lokales Testing)
│   │
│   ├── domain/
│   │   ├── mod.rs
│   │   ├── snapshot.rs                 # Snapshot aggregate
│   │   ├── manifest.rs                 # TOML serialization/deserialization
│   │   ├── package.rs                  # Package, AppSource, PackageSource
│   │   ├── repository.rs              # Repository value object
│   │   ├── config_file.rs             # ConfigFile value object
│   │   ├── mac_default.rs             # MacDefault value object
│   │   └── sensitivity.rs             # Sensitivity levels (public, sensitive, secret)
│   │
│   ├── scanner/
│   │   ├── mod.rs                      # Scanner trait + ScannerRegistry
│   │   ├── trait.rs                    # Scanner trait definition
│   │   │
│   │   ├── packages/                   # Package Managers & Runtimes
│   │   │   ├── mod.rs
│   │   │   ├── homebrew.rs
│   │   │   ├── node.rs                 # nvm/fnm + npm/pnpm globals
│   │   │   ├── python.rs              # pyenv/uv + pip globals
│   │   │   ├── rust_lang.rs           # rustup + cargo
│   │   │   ├── java.rs                # sdkman
│   │   │   ├── flutter.rs             # Flutter/Dart
│   │   │   ├── go.rs
│   │   │   ├── ruby.rs
│   │   │   ├── deno.rs
│   │   │   ├── bun.rs
│   │   │   └── asdf.rs                # asdf/mise universal version manager
│   │   │
│   │   ├── shell/                      # Shell & Terminal
│   │   │   ├── mod.rs
│   │   │   ├── config.rs              # .zshrc, .zshenv, etc.
│   │   │   ├── framework.rs           # oh-my-zsh, prezto
│   │   │   ├── prompt.rs              # starship, p10k
│   │   │   ├── terminal.rs            # iTerm2, Warp, Alacritty
│   │   │   ├── tmux.rs
│   │   │   └── direnv.rs
│   │   │
│   │   ├── git/                        # Git & Version Control
│   │   │   ├── mod.rs
│   │   │   ├── config.rs              # .gitconfig, signing, templates
│   │   │   ├── repos.rs               # Repository discovery + manifest
│   │   │   └── github_cli.rs          # gh CLI config + extensions
│   │   │
│   │   ├── editors/                    # Editors & IDEs
│   │   │   ├── mod.rs
│   │   │   ├── vscode.rs              # VSCode + snippets + profiles
│   │   │   ├── cursor.rs
│   │   │   ├── jetbrains.rs           # IntelliJ, PyCharm, Android Studio
│   │   │   ├── neovim.rs
│   │   │   └── xcode.rs
│   │   │
│   │   ├── containers/                 # Container & Virtualization
│   │   │   ├── mod.rs
│   │   │   └── docker.rs              # Docker, Colima, OrbStack
│   │   │
│   │   ├── cloud/                      # Cloud & DevOps
│   │   │   ├── mod.rs
│   │   │   ├── aws.rs
│   │   │   ├── gcp.rs
│   │   │   ├── kubernetes.rs
│   │   │   └── generic.rs             # Vercel, Fly, Firebase, Terraform etc.
│   │   │
│   │   ├── system/                     # macOS System
│   │   │   ├── mod.rs
│   │   │   ├── defaults.rs            # Dock, Finder, Keyboard, Trackpad, Screenshots, Menu Bar
│   │   │   ├── locale.rs              # Sprache, Region, Timezone, ComputerName
│   │   │   ├── network.rs             # Wi-Fi, VPN, DNS, Proxy
│   │   │   ├── apps.rs                # App Store via mas
│   │   │   ├── login_items.rs
│   │   │   ├── hosts.rs               # /etc/hosts
│   │   │   ├── fonts.rs
│   │   │   └── scheduled.rs           # crontab + LaunchAgents
│   │   │
│   │   ├── security/                   # Security & Credentials
│   │   │   ├── mod.rs
│   │   │   ├── ssh.rs
│   │   │   └── gpg.rs
│   │   │
│   │   ├── productivity/               # Power User Tools
│   │   │   ├── mod.rs
│   │   │   ├── raycast.rs
│   │   │   ├── karabiner.rs
│   │   │   ├── window_manager.rs      # Rectangle, Magnet, BetterTouchTool
│   │   │   ├── browser.rs             # Default Browser, Extensions-Liste
│   │   │   └── api_tools.rs           # Postman, ngrok, mkcert
│   │   │
│   │   ├── ai_tools/                   # AI Developer Tools
│   │   │   ├── mod.rs
│   │   │   ├── claude_code.rs         # ~/.claude/ Config
│   │   │   └── ollama.rs              # Modell-Liste
│   │   │
│   │   ├── xdg_config.rs              # Generic ~/.config/ catch-all scanner
│   │   ├── registries.rs              # .npmrc, pip.conf, cargo config
│   │   ├── databases.rs               # pgpass, psqlrc, TablePlus
│   │   ├── workspace.rs               # Folders + env files
│   │   └── env_files.rs               # .env file collection (encrypted)
│   │
│   ├── bundler/
│   │   ├── mod.rs
│   │   ├── collector.rs               # Sammelt alle Config-Files in staging dir
│   │   ├── dmg.rs                     # DMG creation via hdiutil
│   │   ├── restore_script.rs          # Shell-Script Generator (Handlebars)
│   │   └── encryption.rs              # age encryption wrapper
│   │
│   └── util/
│       ├── mod.rs
│       ├── command.rs                  # Shell command execution helpers
│       ├── progress.rs                 # Progress bar / spinner helpers
│       ├── detection.rs                # Tool detection (is_installed, find_version)
│       └── path.rs                     # Path expansion, home dir helpers
│
├── templates/
│   ├── install.command.hbs             # Handlebars template für Restore-Script
│   ├── stages/                         # Sub-Templates pro Restore-Stage
│   │   ├── homebrew.sh.hbs
│   │   ├── shell.sh.hbs
│   │   ├── runtimes.sh.hbs
│   │   └── ...
│   └── README.md.hbs                  # Template für DMG-README
│
├── config/
│   └── known_xdg_tools.toml           # Registry bekannter ~/.config/ Tools
│
└── tests/
    ├── scanner/
    │   ├── homebrew_test.rs
    │   ├── shell_test.rs
    │   └── ...
    └── integration/
        ├── snapshot_test.rs
        └── manifest_roundtrip_test.rs
```

---

## 11. Dependencies (Cargo.toml)

```toml
[dependencies]
clap = { version = "4", features = ["derive"] }
serde = { version = "1", features = ["derive"] }
toml = "0.8"
walkdir = "2"
regex = "1"
indicatif = "0.17"          # Progress bars
dialoguer = "0.11"           # Interactive prompts
anyhow = "1"                 # Error handling
thiserror = "1"              # Custom errors
handlebars = "5"             # Template engine für restore script
chrono = "0.4"               # Timestamps
dirs = "5"                   # Home directory etc.
which = "6"                  # Find executables in PATH
tracing = "0.1"              # Logging
tracing-subscriber = "0.3"
```

---

## 12. Entwicklungs-Phasen

### Phase 1 — Foundation (MVP)
- [ ] Rust-Projekt aufsetzen (cargo init)
- [ ] CLI-Skeleton mit clap (snapshot, scan, restore)
- [ ] Scanner-Trait definieren
- [ ] Homebrew-Scanner implementieren
- [ ] Shell-Config-Scanner implementieren
- [ ] TOML-Manifest Serialisierung
- [ ] Einfaches Shell-Script generieren (noch kein DMG)
- [ ] Erster End-to-End Test: Scan → Manifest → Script

### Phase 2 — Core Scanners
- [ ] Git-Repos Scanner
- [ ] Node.js Scanner (nvm/fnm + globals)
- [ ] Python Scanner (pyenv/uv + globals)
- [ ] Rust Scanner (rustup + cargo)
- [ ] VSCode/Cursor Scanner

### Phase 3 — System Scanners
- [ ] macOS Defaults Scanner
- [ ] App Store Scanner (mas)
- [ ] Fonts Scanner
- [ ] LaunchAgents/Cron Scanner
- [ ] Folder Structure Scanner

### Phase 4 — Security & Bundling
- [ ] age-Encryption für SSH Keys + .env Files
- [ ] DMG-Erstellung via hdiutil
- [ ] Interactive Mode (--interactive)
- [ ] Dry-Run Mode (--dry-run)

### Phase 5 — Polish
- [ ] Restore-Script mit Progress-Ausgabe
- [ ] Idempotenz in allen Restore-Stages
- [ ] Error Recovery (einzelne Stage fails → weiter)
- [ ] README im DMG generieren
- [ ] Architecture-Check (Intel vs ARM Warnings)

---

## 13. Sensitivity Model

Nicht alle Daten sind gleich. machinist kategorisiert alles in drei Stufen:

| Level | Beschreibung | Beispiele | Handling |
|---|---|---|---|
| **public** | Keine Sicherheitsbedenken | Homebrew Packages, VSCode Extensions, macOS Defaults | Klartext im Manifest |
| **sensitive** | Könnte private Infos enthalten | .npmrc (Registry-Auth), AWS config (Account-IDs), .gitconfig (Email) | Im Manifest, aber User wird gewarnt |
| **secret** | Credentials, Keys, Tokens | SSH Keys, .env Files, .pgpass, Docker Auth | NUR mit age-Encryption, explicit opt-in |

Beim `--interactive` Scan wird der User pro Kategorie gefragt:
```
Found 3 secret items:
  🔑 ~/.ssh/id_ed25519          [SSH Private Key]
  🔑 ~/Code/app/.env            [Environment File]
  🔑 ~/.pgpass                  [PostgreSQL Passwords]

Include these? They will be encrypted with a passphrase. [y/N]
```

---

## 14. Offene Fragen / Entscheidungen

1. **Version Pinning**: Sollen Packages mit exakter Version installiert werden oder immer `latest`?
   → Vorschlag: Default `latest`, optional `--pin-versions`

2. **Private Repos**: Wie handeln wir Auth auf neuem Mac?
   → SSH Key Restore (Stage 2) passiert VOR Git Clone (Stage 11) — genau deswegen

3. **Große Repos**: Shallow Clone als Default?
   → Vorschlag: Option pro Repo im Manifest (`shallow = true/false`), Default: full clone

4. **Cursor vs VSCode**: Separater Scanner oder gemeinsamer?
   → Vorschlag: Getrennte Scanner (verschiedene Config-Pfade), aber shared Logic

5. **Config Conflicts**: Was wenn auf neuem Mac schon eine .zshrc existiert?
   → Vorschlag: Backup als `.zshrc.machinist-backup` + Timestamp vor Überschreiben

6. **Intel ↔ Apple Silicon**: Manche Casks haben unterschiedliche Binaries
   → Vorschlag: Arch im Manifest speichern, beim Restore warnen wenn Mismatch

7. **Homebrew Bundle vs eigene Logik**: Homebrew hat `brew bundle` — nutzen oder eigenes?
   → Vorschlag: Eigene Logik, da wir mehr Kontrolle und Progress-Reporting brauchen

8. **Partial Restore**: User will nur Shell + Git Setup, nicht den ganzen Stack
   → Vorschlag: `machinist restore --only shell,git,ssh` oder `--skip cloud,docker`

9. **Update/Re-Scan**: User hat seit dem Snapshot neue Tools installiert
   → Vorschlag: `machinist snapshot --update` merged mit bestehendem Manifest (v2 Feature)

10. **Multi-Machine Profiles**: Ein Dev hat Work-Mac und Private-Mac
    → Vorschlag: Manifest-Profiles oder getrennte Snapshots (v2 Feature)

11. **Ollama Modell-Downloads**: Modelle können mehrere GB groß sein — in Stage 29 neben Docker Images?
    → Vorschlag: Eigene optionale Stage, parallel zu Docker pulls

12. **Browser-Bookmarks**: Export/Import von Bookmarks (Chrome/Firefox haben JSON/HTML-Export)?
    → Vorschlag: Optional, explicit opt-in — Bookmarks können sehr persönlich sein

13. **Cloud-Upload des Snapshots**: DMG nur lokal oder auch Upload zu S3/iCloud/NAS?
    → Vorschlag: v1 nur lokal, v2 optional `machinist snapshot --upload s3://bucket/`
