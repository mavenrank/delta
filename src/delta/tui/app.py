from __future__ import annotations

import os
import time
from datetime import datetime
from typing import Optional

from textual.app import App, ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Label, Static
from textual.binding import Binding

from delta.config import Config
from delta.git_info import Health, Info
from delta.scanner import Repo, scan_folders_with_git_info


class DeltaApp(App):
    CSS_PATH = "styles.tcss"

    BINDINGS = [
        Binding("q", "quit", "quit"),
        Binding("r", "refresh", "refresh"),
        Binding("a", "add_folder", "add folder"),
        Binding("/", "filter", "filter"),
        Binding("escape", "cancel_input", "cancel"),
    ]

    def __init__(self, config: Config, config_path: str) -> None:
        super().__init__()
        self.config = config
        self.config_path = config_path
        self.repos: list[Repo] = []
        self.filtered: list[Repo] = []
        self.filtering = False
        self.filter_text = ""
        self.adding = False
        self.add_text = ""
        self.scan_time: float = 0.0
        self.error: Optional[str] = None

    def compose(self) -> ComposeResult:
        yield Static("delta  -  repo scanner v0.2.0", id="header")
        yield DataTable(id="repo-table", cursor_type="row")
        yield Static(self._render_footer(), id="footer")

    def on_mount(self) -> None:
        table = self.query_one("#repo-table", DataTable)
        table.add_column("Repo", key="repo", width=28)
        table.add_column("Branch", key="branch", width=14)
        table.add_column("St", key="status", width=4)
        table.add_column("Health", key="health", width=12)
        table.add_column("Remote", key="remote", width=18)
        table.add_column("Last Commit", key="last_commit", width=16)
        table.add_column("Path", key="path", width=40)
        self._do_scan()

    def _do_scan(self) -> None:
        self.error = None
        table = self.query_one("#repo-table", DataTable)
        table.clear()
        footer = self.query_one("#footer", Static)
        footer.update("  Scanning folders...")

        start = time.time()
        try:
            self.repos = scan_folders_with_git_info(
                [os.path.expanduser(f) for f in self.config.scan_folders]
            )
        except Exception as e:
            self.error = str(e)
            self.repos = []

        self.scan_time = time.time() - start
        self.filtered = self.repos
        self._refresh_table()

    def _refresh_table(self) -> None:
        table = self.query_one("#repo-table", DataTable)
        table.clear()

        name_counts: dict[str, int] = {}
        for repo in self.repos:
            name_counts[repo.name] = name_counts.get(repo.name, 0) + 1

        for repo in self.filtered:
            display_name = repo.name
            if name_counts.get(repo.name, 0) > 1:
                parent = os.path.basename(os.path.dirname(repo.path))
                display_name = f"{repo.name} ({parent})"

            row_data = self._build_row(repo, display_name)
            table.add_row(*row_data, key=repo.path)

        footer = self.query_one("#footer", Static)
        footer.update(self._render_footer())

    def _build_row(self, repo: Repo, display_name: str) -> tuple:
        name = self._truncate(display_name, 27)

        if repo.git_info:
            gi = repo.git_info

            if gi.detached:
                branch = "(detached)"
            else:
                branch = self._truncate(gi.branch, 13)

            if gi.status.is_clean:
                status = "v"
            else:
                status = "M"

            health = str(gi.health)

            remote = gi.remote_summary()

            if gi.last_commit:
                last_commit = gi.last_commit.relative_time()
                if gi.last_commit.is_stale(self.config.stale.threshold_days):
                    last_commit = f"! {last_commit}"
            else:
                last_commit = "-"
        else:
            branch = "-"
            status = "-"
            health = "-"
            remote = "no git"
            last_commit = "-"

        path = self._shorten_path(repo.path)

        return (name, branch, status, health, remote, last_commit, path)

    def _shorten_path(self, p: str) -> str:
        home = os.path.expanduser("~")
        if p.startswith(home):
            p = "~" + p[len(home):]
        p = p.replace("\\", "/")
        return p

    def _truncate(self, s: str, n: int) -> str:
        if len(s) <= n:
            return s
        return s[: n - 1] + "..."

    def _render_footer(self) -> str:
        parts = [f"{len(self.filtered)} repos"]
        if self.filter_text and not self.filtering:
            parts.append(f"filtered from {len(self.repos)}")
        if self.scan_time > 0:
            parts.append(f"scan: {self.scan_time:.3f}s")
        if self.error:
            parts.append(f"error: {self.error}")

        stats = "  -  ".join(parts)
        keys = "  [up/down] navigate  [r] refresh  [/] filter  [a] add folder  [q] quit"

        result = f"  {stats}\n{keys}"

        if self.filtering:
            result += f"\n  / {self.filter_text}  (Enter to apply, Esc to cancel)"
        if self.adding:
            result += f"\n  add folder: {self.add_text}  (Enter to save, Esc to cancel)"

        return result

    def action_refresh(self) -> None:
        self._do_scan()

    def action_filter(self) -> None:
        self.filtering = True
        self.filter_text = ""
        footer = self.query_one("#footer", Static)
        footer.update(self._render_footer())

    def action_add_folder(self) -> None:
        self.adding = True
        self.add_text = ""
        footer = self.query_one("#footer", Static)
        footer.update(self._render_footer())

    def action_cancel_input(self) -> None:
        self.filtering = False
        self.filter_text = ""
        self.adding = False
        self.add_text = ""
        self.filtered = self.repos
        self._refresh_table()

    def on_key(self, event) -> None:
        if self.filtering:
            self._handle_filter_key(event)
        elif self.adding:
            self._handle_add_key(event)

    def _handle_filter_key(self, event) -> None:
        event.prevent_default()
        event.stop()
        key = event.key

        if key == "enter":
            self.filtering = False
            self._apply_filter()
        elif key == "backspace":
            if self.filter_text:
                self.filter_text = self.filter_text[:-1]
                self._apply_filter()
        elif key == "escape":
            self.filtering = False
            self.filter_text = ""
            self.filtered = self.repos
            self._refresh_table()
        elif len(key) == 1 and key.isprintable():
            self.filter_text += key
            self._apply_filter()

        footer = self.query_one("#footer", Static)
        footer.update(self._render_footer())

    def _handle_add_key(self, event) -> None:
        event.prevent_default()
        event.stop()
        key = event.key

        if key == "enter":
            path = self.add_text.strip().strip('"')
            if path:
                try:
                    from delta import config as cfg_mod
                    cfg_mod.add_folder(self.config, path)
                    cfg_mod.save(self.config, self.config_path)
                    self.adding = False
                    self.add_text = ""
                    self._do_scan()
                    return
                except ValueError as e:
                    self.error = str(e)
        elif key == "backspace":
            if self.add_text:
                self.add_text = self.add_text[:-1]
        elif key == "escape":
            self.adding = False
            self.add_text = ""
        elif len(key) == 1 and key.isprintable():
            self.add_text += key

        footer = self.query_one("#footer", Static)
        footer.update(self._render_footer())

    def _apply_filter(self) -> None:
        if not self.filter_text:
            self.filtered = self.repos
        else:
            lower = self.filter_text.lower()
            self.filtered = [r for r in self.repos if lower in r.name.lower()]
        self._refresh_table()
