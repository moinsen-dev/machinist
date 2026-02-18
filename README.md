# machinist

**Mac Developer Environment Snapshot & Restore CLI**

Scan. Bundle. Restore. Identical dev setup in minutes.

---

A developer buys a new Mac. Instead of spending days installing tools, copying shell configs, and cloning repos, they run **one DMG** — and have an identical working environment in 20 minutes.

**machinist** is a Go CLI tool that:

1. **Scans** your current Mac (dev tools, system settings, repos, configs)
2. **Generates** a structured TOML manifest of everything found
3. **Bundles** the manifest + config files into a DMG
4. **Restores** everything on a new Mac via a generated shell script

machinist is **not a backup tool** — it's a setup automator. The DMG contains instructions (`brew install`, `git clone`, `mkdir -p`), not copies of your tools. Config files that can't be installed via a package manager are included directly.

## How It Works

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

## Usage

```bash
# Snapshot — capture your existing system
machinist snapshot
machinist snapshot --output ~/Desktop/machinist.dmg
machinist snapshot --interactive
machinist snapshot --dry-run

# Scan — run individual scanners
machinist scan homebrew
machinist scan shell
machinist scan git-repos --search-paths ~/Code,~/Projects

# Compose — build a setup from a profile (AI or manual)
machinist compose --from profile://flutter-ios
machinist compose --from profile://fullstack-js --add docker,postgres
machinist compose --from manifest.toml --output setup.dmg

# Restore — on a new Mac, from mounted DMG
machinist restore
machinist restore --skip homebrew,fonts
machinist restore --dry-run
machinist restore --only shell,git,ssh

# MCP Server — let AI tools drive machinist
machinist serve                           # stdio (Claude Code, Cursor)
machinist serve --port 3333               # SSE (Claude Desktop, web clients)

# Info
machinist list-scanners
machinist list-profiles
machinist version
```

## AI-Powered Setup

machinist includes an **MCP (Model Context Protocol) server**, allowing any MCP-compatible AI to compose and build Mac setups interactively.

```
User → AI: "I'm a Flutter developer, I work with Firebase and Docker,
            and I use VS Code."

AI calls machinist MCP tools:
  1. list_profiles()              → finds "flutter-ios" as base
  2. compose_manifest(            → adds Docker, Firebase, VS Code extensions
       base: "flutter-ios",
       add: ["docker", "firebase"])
  3. build_dmg(manifest)          → generates the DMG

User runs DMG on new Mac → done.
```

Configure in Claude Code or Claude Desktop:

```json
{
  "mcpServers": {
    "machinist": {
      "command": "machinist",
      "args": ["serve"]
    }
  }
}
```

### Built-in Profiles

| Profile | Description |
|---|---|
| `minimal` | Homebrew, Git, zsh+starship, VS Code |
| `fullstack-js` | Node.js, pnpm, Docker, Postgres, VS Code |
| `flutter-ios` | Flutter, Xcode, Simulators, CocoaPods |
| `python-data` | Python, Jupyter, conda, Docker |
| `rust-dev` | Rust, rustup+nightly, cargo tools |
| `go-dev` | Go, Docker, Postgres, VS Code |
| `devops` | Docker, Kubernetes, Terraform, AWS/GCP CLI |

Profiles are composable — use them as a base and add/remove packages as needed.

## What Gets Captured

machinist uses modular **scanners**, each responsible for a specific domain. Only scanners that find results produce output in the manifest.

### Package Managers & Runtimes
Homebrew (taps, formulae, casks, services), Node.js (nvm/fnm + globals), Python (pyenv/uv + globals), Rust (rustup + cargo), Ruby, Go, Java/Kotlin (SDKMAN), Flutter/Dart, Deno, Bun, asdf/mise

### Shell & Terminal
Shell config files (.zshrc, .zshenv, etc.), frameworks (oh-my-zsh, prezto), prompts (starship, p10k), terminal emulators (iTerm2, Warp, Alacritty, WezTerm), tmux, direnv

### Git & Version Control
Git config, global gitignore, templates, signing keys (GPG/SSH), credential helpers, GitHub CLI (config + extensions), all git repos (remote URLs, branches, submodules)

### Editors & IDEs
VSCode, Cursor, JetBrains (IntelliJ, PyCharm, WebStorm, Android Studio), Neovim, Vim, Xcode

### Containers & Virtualization
Docker (config, frequently used images), Colima, OrbStack, Podman

### Cloud & DevOps
AWS CLI, GCP, Azure, Kubernetes, Terraform, Vercel, Fly.io, Firebase, Cloudflare

### macOS System
Dock, Finder, Keyboard, Trackpad, Mission Control, Accessibility, Login Items, Hosts file, Spotlight, Screenshots, Locale/Timezone, Display, Menu Bar, Computer Name

### Network & VPN
Wi-Fi networks, VPN configurations, DNS settings, Proxy settings, Tailscale

### Security & Credentials
SSH keys (encrypted with [age](https://github.com/FiloSottile/age)), GPG keys (encrypted), age keys

### And More
- **Fonts** — user-installed fonts + Nerd Fonts via Homebrew
- **Productivity tools** — Raycast, Alfred, Karabiner-Elements, Rectangle, BetterTouchTool, 1Password CLI
- **Database clients** — PostgreSQL, MySQL, TablePlus, DBeaver, Redis
- **Package registries** — npm, pip, Cargo, Gem, CocoaPods, Pub
- **Browser** — default browser, extensions checklist
- **AI developer tools** — Claude Code, Ollama models
- **API tools** — Postman/Insomnia exports, ngrok, mkcert
- **XDG config catch-all** — bat, lazygit, htop, btop, aerospace, and custom paths
- **Workspace** — folder structure, git repos, .env files (encrypted)
- **Scheduled tasks** — crontab, LaunchAgents

## Restore Order

The restore script executes 31 stages in dependency order — SSH keys are decrypted before git repos are cloned, Homebrew is installed before language runtimes, etc. Each stage is:

- **Skippable** (`--skip stage-name`)
- **Idempotent** (checks if already installed/present)
- **Logged** to `~/.machinist/restore.log`
- **Fault-tolerant** (logs errors, continues to next stage)

A **post-restore checklist** is generated for things that can't be automated: macOS permissions (TCC), browser extensions, Bluetooth pairing, VPN passwords, etc.

## Security

| Concern | Solution |
|---|---|
| SSH keys | Encrypted with [age](https://github.com/FiloSottile/age) — set passphrase during snapshot, enter during restore |
| .env files | Same age encryption |
| Sensitive defaults | Scanner asks explicitly ("Include SSH keys? [y/N]") |
| DMG password | Optional: encrypt DMG itself via `hdiutil` |

Data is categorized into three sensitivity levels:

- **public** — Homebrew packages, VSCode extensions, macOS defaults (plaintext in manifest)
- **sensitive** — .npmrc, AWS config, .gitconfig (included with user warning)
- **secret** — SSH keys, .env files, .pgpass (age-encrypted, explicit opt-in only)

## Architecture

- **Go** for scanner + bundler — fast compilation, excellent `os/exec` for shell commands, single binary, `text/template` in stdlib
- **Shell script** for restore — must run on vanilla Mac without Go
- **TOML** for manifest — human-readable, human-editable (BurntSushi/toml)
- **filippo.io/age** for encryption — the reference implementation, written in Go
- **hdiutil** for DMG — macOS-native, no extra dependency
- **cobra** for CLI — de-facto standard for Go CLIs (used by kubectl, gh, docker)
- **text/template** (stdlib) for restore script generation

## Development

```bash
# Build
go build ./cmd/machinist

# Run
go run ./cmd/machinist snapshot --dry-run
go run ./cmd/machinist scan homebrew
go run ./cmd/machinist list-scanners

# Test
go test ./...
```

## License

[MIT](LICENSE)
