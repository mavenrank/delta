# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned for v0.2.0 — Core TUI + Git Awareness
- bubbletea + lipgloss TUI framework
- Interactive table with repo status
- Git branch, dirty/clean, last commit info
- Health indicators and stale detection
- Keyboard navigation, filtering, folder management

### Planned for v0.3.0 — Full Feature Set
- Multi-remote support with ahead/behind counts
- Groups/tags system with auto-assignment
- Detail panel with full repo info
- Open in editor, sortable columns
- Watch mode, parallel scanning, error handling

## [0.1.0] - 2026-06-20

### Added
- Project scaffold with Go module structure
- JSON config reader/writer for scan folders and settings
- Recursive folder scanner (finds `.git` directories at any depth)
- Non-git code folder detection
- Interactive folder-add prompt (handles quoted paths)
- `VERSION.txt`, `README.md`, `LICENSE` (MIT), `.gitignore`
- `CHANGELOG.md` and `ROADMAP.md` for tracking

[Unreleased]: https://codeberg.org/mavenrank/delta/compare/v0.1.0...HEAD
[0.1.0]: https://codeberg.org/mavenrank/delta/releases/tag/v0.1.0
