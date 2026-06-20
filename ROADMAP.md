# Roadmap

Planned features for future versions of `delta`.

## v0.2.0 — "Basic Table"
- [ ] bubbletea + lipgloss TUI framework
- [ ] Table with columns: Repo Name, Branch, Status, Path
- [ ] Color-coded status (clean / dirty / untracked)
- [ ] Keyboard navigation (arrows, q quit, r refresh)
- [ ] Basic status bar (repo count, scan time)

## v0.3.0 — "Git Awareness"
- [ ] Git branch detection
- [ ] Dirty working tree parsing (modified/untracked/staged counts)
- [ ] Last commit message + relative time
- [ ] Health indicators: clean / ahead / behind / diverged / dirty / detached HEAD
- [ ] Stale detection (last commit >90 days → warning flag)

## v0.4.0 — "Multi-Remote"
- [ ] Parse all remotes from `.git/config`
- [ ] Show remote names + URLs
- [ ] Ahead/behind counts per remote
- [ ] Local-only branch warning (no upstream)
- [ ] Branch age detection

## v0.5.0 — "Groups & Organization"
- [ ] Groups/tags system in config
- [ ] Auto-assign groups by folder pattern
- [ ] Manual group override in config
- [ ] Group filter dropdown (g key)
- [ ] Group column in table
- [ ] Per-group color coding

## v0.6.0 — "Interactive Detail"
- [ ] Detail panel on Enter (full repo info)
- [ ] Shows: path, all remotes with ahead/behind, full git status, last commit, branch age, health
- [ ] Open in VS Code or configured editor
- [ ] Filter/search by name (/ key)
- [ ] Sortable columns

## v0.7.0 — "Polish"
- [ ] Watch/auto-refresh mode (-w flag or toggle)
- [ ] Parallel scanning with goroutines
- [ ] Error handling edge cases (deleted repos, permission errors, broken .git)
- [ ] Custom config path flag (-c)
- [ ] Color scheme customization
- [ ] Performance optimization

## v1.0.0 — "Stable"
- [ ] Full test suite
- [ ] Cross-platform testing (Windows/Linux/Mac)
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Binary releases
- [ ] Complete documentation
