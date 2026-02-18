# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Project architecture document (`pad.md`) with complete domain model
- Scanner specifications for 19 categories covering the full Mac developer environment
- Manifest format (TOML) with all scanner sections
- Restore script architecture with 31 ordered stages
- Security model with three sensitivity levels (public, sensitive, secret)
- DMG bundle structure specification
- Post-restore checklist for non-automatable steps (TCC permissions, Bluetooth, browser extensions)
- README with usage examples and feature overview
- MIT license
- This changelog
- MCP (Model Context Protocol) server architecture — AI tools can drive machinist
- MCP tools: list_scanners, scan, compose_manifest, validate_manifest, build_dmg, diff_manifests
- MCP resources: system/snapshot, profiles
- Profile system with 10 built-in presets (minimal, fullstack-js, flutter-ios, python-data, etc.)
- `machinist compose` command for building setups from profiles
- `machinist serve` command for running as MCP server (stdio + SSE)

### Changed
- Switched from Rust to Go (better fit for shell-command orchestration, age reference impl in Go, faster dev velocity)
- Go project structure with `cmd/`, `internal/`, `mcp/`, `profiles/` layout
- Dependencies: cobra, BurntSushi/toml, filippo.io/age, bubbletea, mcp-go
- Development phases reorganized: Phase 5 is now MCP Server & Profiles, Phase 6 is Polish
