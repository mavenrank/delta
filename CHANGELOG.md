# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned for v0.3.0 — Full Feature Set
- Multi-remote support with ahead/behind counts per remote
- Groups/tags system with auto-assignment by folder pattern
- Detail panel with full repo info (Enter key)
- Open in VS Code or configured editor
- Sortable columns
- Watch/auto-refresh mode
- Parallel scanning
- Error handling edge cases
- Local-only branch warning (no upstream)
- Branch age detection

## [0.2.0] - 2026-06-21

### Changed
- **Ported from Go to Python** (Textual + Typer + uv)
- Config format changed from JSON to YAML with pydantic validation
- CLI now uses Typer with subcommands: `delta tui`, `delta scan`, `delta config`
- TUI rebuilt with Textual DataTable (fixes alignment, scrolling, footer jumping)
- Config now supports column customization and stale threshold settings

### Added
- `delta config show` — display current configuration
- `delta config add <path>` — add a scan folder from CLI
- `delta config path` — show config file location
- `delta config init` — create a default config file
- `delta version` — show version
- `delta scan --detail` — scan with git details (branch, health, remote)
- YAML config with columns and stale sections
- Mouse/trackpad scroll support (built-in via Textual)
- Resizable columns (Textual DataTable)
- Fixed footer via CSS dock layout
- Duplicate repo name disambiguation with parent folder

### Removed
- Go implementation (main.go, internal/, go.mod, go.sum)
- JSON config format (replaced by YAML)
- Manual scroll range calculation (Textual handles natively)
- Manual column padding (Textual DataTable handles natively)

### Fixed
- Selected row no longer shifts alignment (Textual manages row rendering)
- Footer stays fixed at bottom (CSS dock layout)
- No more index out of range panic (Textual handles scrolling)
- Filter mode properly exits on Enter/Esc
- Ctrl/CapsLock no longer adds junk characters in filter mode

## [0.1.0] - 2026-06-20

### Added
- Project scaffold with Go module structure
- JSON config reader/writer for scan folders and settings
- Recursive folder scanner (finds `.git` directories at any depth)
- Non-git code folder detection
- Interactive folder-add prompt (handles quoted paths)
- `VERSION.txt`, `README.md`, `LICENSE` (MIT), `.gitignore`
- `CHANGELOG.md` and `ROADMAP.md` for tracking

[Unreleased]: https://codeberg.org/mavenrank/delta/compare/v0.2.0...HEAD
[0.2.0]: https://codeberg.org/mavenrank/delta/releases/tag/v0.2.0
[0.1.0]: https://codeberg.org/mavenrank/delta/releases/tag/v0.1.0
