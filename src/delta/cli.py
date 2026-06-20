from __future__ import annotations

import os
import sys
from typing import Optional

import typer

from delta import __version__
from delta import config as cfg_mod
from delta.scanner import scan_folders, scan_folders_with_git_info

app = typer.Typer(
    name="delta",
    help="A terminal UI tool for scanning and managing local Git repositories.",
    no_args_is_help=True,
)

config_app = typer.Typer(name="config", help="Manage delta configuration.")
app.add_typer(config_app, name="config")


@app.command()
def tui(
    config_path: str = typer.Option("", "--config", "-c", help="Path to config file"),
) -> None:
    """Open the interactive TUI."""
    cfg = cfg_mod.load(config_path)
    resolved = config_path or cfg_mod.default_path()

    from delta.tui.app import DeltaApp

    delta_app = DeltaApp(cfg, resolved)
    delta_app.run()


@app.command()
def scan(
    config_path: str = typer.Option("", "--config", "-c", help="Path to config file"),
    detail: bool = typer.Option(False, "--detail", "-d", help="Show git details"),
) -> None:
    """Scan folders and print results."""
    cfg = cfg_mod.load(config_path)

    if not cfg.scan_folders:
        typer.echo("No scan folders configured. Use 'delta config add <path>' to add folders.")
        raise typer.Exit(1)

    folders = [os.path.expanduser(f) for f in cfg.scan_folders]

    if detail:
        repos = scan_folders_with_git_info(folders)
    else:
        repos = scan_folders(folders)

    typer.echo(f"found {len(repos)} repos")
    for repo in repos:
        if detail and repo.git_info:
            gi = repo.git_info
            typer.echo(
                f"  {repo.name:30s}  branch={gi.branch:12s}  "
                f"health={str(gi.health):10s}  remote={gi.remote_summary():20s}  "
                f"({repo.path})"
            )
        else:
            typer.echo(f"  {repo.name} ({repo.path})")


@app.command()
def version() -> None:
    """Show version information."""
    typer.echo(f"delta v{__version__}")


@config_app.command("show")
def config_show(
    config_path: str = typer.Option("", "--config", "-c", help="Path to config file"),
) -> None:
    """Show current configuration."""
    cfg = cfg_mod.load(config_path)
    resolved = config_path or cfg_mod.default_path()
    typer.echo(f"Config file: {resolved}")
    typer.echo(f"Scan folders ({len(cfg.scan_folders)}):")
    for folder in cfg.scan_folders:
        typer.echo(f"  - {folder}")
    typer.echo(f"Editor: {cfg.editor}")
    typer.echo(f"Path mode: {cfg.columns.path}")
    typer.echo(f"Stale detection: {'on' if cfg.stale.enabled else 'off'} ({cfg.stale.threshold_days} days)")


@config_app.command("add")
def config_add(
    folder: str = typer.Argument(..., help="Folder path to add to scan list"),
    config_path: str = typer.Option("", "--config", "-c", help="Path to config file"),
) -> None:
    """Add a scan folder to the configuration."""
    cfg = cfg_mod.load(config_path)
    resolved = config_path or cfg_mod.default_path()
    try:
        cfg_mod.add_folder(cfg, folder)
        cfg_mod.save(cfg, resolved)
        typer.echo(f"Added: {folder}")
    except ValueError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(1)


@config_app.command("path")
def config_path_cmd(
    config_path: str = typer.Option("", "--config", "-c", help="Path to config file"),
) -> None:
    """Show the config file path."""
    resolved = config_path or cfg_mod.default_path()
    typer.echo(resolved)


@config_app.command("init")
def config_init(
    config_path: str = typer.Option("", "--config", "-c", help="Path to config file"),
) -> None:
    """Create a default config file if it doesn't exist."""
    resolved = config_path or cfg_mod.default_path()
    if os.path.exists(resolved):
        typer.echo(f"Config already exists: {resolved}")
    else:
        cfg = cfg_mod.Config()
        cfg_mod.save(cfg, resolved)
        typer.echo(f"Created: {resolved}")
