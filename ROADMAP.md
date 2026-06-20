# Roadmap

Planned features for future versions of `delta`.

## v0.2.0 — Core TUI + Git Awareness
- [ ] bubbletea + lipgloss TUI framework
- [ ] Table with columns: Repo Name, Branch, Status, Path, Last Commit
- [ ] Color-coded status (clean / dirty / untracked)
- [ ] Keyboard navigation (arrows, q quit, r refresh, a add folder)
- [ ] Git branch detection
- [ ] Dirty working tree parsing (modified/untracked/staged counts)
- [ ] Last commit message + relative time
- [ ] Health indicators: clean / ahead / behind / diverged / dirty / detached HEAD
- [ ] Stale detection (last commit >90 days → warning flag)
- [ ] Status bar (repo count, scan time)
- [ ] Add folder via interactive prompt (handles quotes)
- [ ] Filter/search by name (/ key)

## v0.3.0 — Full Feature Set
- [ ] Parse all remotes from `.git/config`
- [ ] Show remote names + URLs
- [ ] Ahead/behind counts per remote (up/down arrows)
- [ ] Local-only branch warning (no upstream)
- [ ] Branch age detection
- [ ] Groups/tags system in config
- [ ] Auto-assign groups by folder pattern
- [ ] Manual group override in config
- [ ] Group filter dropdown (g key)
- [ ] Group column + per-group colors
- [ ] Detail panel on Enter (full repo info)
- [ ] Open in VS Code or configured editor
- [ ] Sortable columns
- [ ] Watch/auto-refresh mode (-w flag or toggle)
- [ ] Parallel scanning with goroutines
- [ ] Error handling edge cases (deleted repos, permission errors, broken .git)
- [ ] Custom config path flag (-c)

## Future (post-v1.0.0)
- [ ] Cross-platform testing (Windows/Linux/Mac)
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Binary releases
- [ ] Complete documentation
- [ ] Full test suite
