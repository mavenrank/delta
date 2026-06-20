# delta

A terminal UI tool for scanning and managing local Git repositories across multiple machines.

## Overview

`delta` recursively scans configured folders for Git repositories and displays their status in an interactive TUI — branch, health, remotes, ahead/behind counts, and more. Designed for developers who manage many repos across multiple remotes (GitHub, Codeberg, local) and multiple machines.

Built with [Go](https://go.dev/) and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

### Current (v0.1.0)
- Recursive folder scanning for Git repos
- JSON config for scan folders and settings
- Non-git code folder detection

### Planned
See [ROADMAP.md](ROADMAP.md) for the full feature roadmap.

## Installation

```sh
go build -o delta.exe
```

## Usage

```sh
./delta.exe
```

## Configuration

Config is stored at `~/.config/delta/config.json`:

```json
{
  "scan_folders": [
    "C:\\Users\\FrostyFjord\\Repos",
    "C:\\Users\\FrostyFjord\\non-git-code"
  ],
  "editor": "code"
}
```

## Versioning

This project uses [Semantic Versioning 2.0.0](https://semver.org/).
See [VERSION.txt](VERSION.txt) for the current version and [CHANGELOG.md](CHANGELOG.md) for release history.

## License

[MIT](LICENSE)
