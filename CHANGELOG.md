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
- Watch/auto-refresh mode (-w flag or toggle)
- Parallel scanning with goroutines
- Error handling edge cases (deleted repos, permission errors, broken .git)
- Custom config path flag (-c)
- Local-only branch warning (no upstream)
- Branch age detection

## [0.2.0] - 2026-06-20

### Added
- bubbletea + lipgloss TUI framework with alt-screen rendering
- Interactive table with columns: Repo Name, Branch, Status, Health, Last Commit, Path
- Color-coded status indicators (clean / dirty / untracked)
- Health indicators: clean, ahead, behind, diverged, dirty, detached HEAD
- Git branch detection via `git rev-parse`
- Dirty working tree parsing (modified/untracked/staged counts)
- Last commit message + relative time display
- Stale repository detection (last commit >90 days → warning flag)
- Status bar with repo count and scan duration
- Keyboard navigation (arrows/j/k, q quit, r refresh)
- Filter/search by name (/ key with live filtering)
- Add folder via interactive prompt (handles quoted paths, saves to config)
- Internal `git` package for running and parsing git commands

### Changed
- Scanner now returns rich Repo structs with GitInfo
- Main launches TUI instead of plain text output (scan-only mode still via -scan flag)
- Config path exported as `DefaultPath()` for external use

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
