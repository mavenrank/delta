# delta

A terminal UI tool for scanning and managing local Git repositories across multiple machines.

## Overview

`delta` recursively scans configured folders for Git repositories and displays their status in an interactive TUI — branch, health, remotes, ahead/behind counts, and more. Designed for developers who manage many repos across multiple remotes (GitHub, Codeberg, local) and multiple machines.

Built with [Python](https://python.org/), [Textual](https://textual.textualize.io/), [Typer](https://typer.tiangolo.com/), and [uv](https://docs.astral.sh/uv/).

## Installation

```sh
uv tool install .
```

This installs `delta` as a global command. Alternatively, run directly from the project folder:

```sh
uv run delta tui
```

## CLI Commands

```sh
delta tui                      # open the interactive TUI (default)
delta scan                     # scan folders and print results
delta scan --detail            # scan with git details (branch, health, remote)
delta config show              # show current configuration
delta config add ~/Repos       # add a scan folder
delta config path              # show config file location
delta config init              # create a default config file
delta version                  # show version
```

## Configuration

Config is stored at `~/.config/delta/config.yaml`:

```yaml
scan_folders:
  - ~/Repos
  - ~/GitHub
  - ~/Codeberg
  - ~/non-git-code

editor: code

columns:
  repo: true
  branch: true
  status: true
  health: true
  remote: true
  last_commit: true
  path: short          # full | short | hidden

stale:
  enabled: true
  threshold_days: 90
```

## TUI Keys

| Key | Action |
|---|---|
| `↑` / `↓` | Navigate repos |
| `r` | Refresh — rescan all folders |
| `/` | Filter repos by name (Enter to apply, Esc to cancel) |
| `a` | Add a scan folder (Enter to save, Esc to cancel) |
| `q` | Quit |

## Versioning

This project uses [Semantic Versioning 2.0.0](https://semver.org/).
See [VERSION.txt](VERSION.txt) for the current version and [CHANGELOG.md](CHANGELOG.md) for release history.

## License

[MIT](LICENSE)
