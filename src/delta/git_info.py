from __future__ import annotations

import subprocess
import time
from dataclasses import dataclass, field
from enum import IntEnum
from typing import Optional


class Health(IntEnum):
    CLEAN = 0
    AHEAD = 1
    BEHIND = 2
    DIVERGED = 3
    DIRTY = 4
    DETACHED = 5
    UNKNOWN = 6

    def __str__(self) -> str:
        names = {
            0: "clean",
            1: "ahead",
            2: "behind",
            3: "diverged",
            4: "dirty",
            5: "detached",
            6: "unknown",
        }
        return names.get(int(self), "unknown")

    @property
    def icon(self) -> str:
        icons = {
            0: "o",
            1: "^",
            2: "v",
            3: "/\\",
            4: "o",
            5: "<>",
            6: "?",
        }
        return icons.get(int(self), "?")


@dataclass
class Status:
    modified: int = 0
    untracked: int = 0
    staged: int = 0
    is_clean: bool = True

    @property
    def icon(self) -> str:
        return "v" if self.is_clean else "M"


@dataclass
class Commit:
    message: str = ""
    date: float = 0.0

    def relative_time(self) -> str:
        if self.date == 0.0:
            return ""
        delta = time.time() - self.date
        if delta < 3600:
            return f"{int(delta // 60)}m ago"
        if delta < 86400:
            return f"{int(delta // 3600)}h ago"
        if delta < 30 * 86400:
            return f"{int(delta // 86400)}d ago"
        if delta < 365 * 86400:
            return f"{int(delta // (30 * 86400))}mo ago"
        return f"{int(delta // (365 * 86400))}y ago"

    def is_stale(self, threshold_days: int = 90) -> bool:
        if self.date == 0.0:
            return False
        return (time.time() - self.date) > threshold_days * 86400


@dataclass
class Remote:
    name: str
    url: str


def _remote_label(url: str) -> str:
    if "github.com" in url:
        return "github"
    if "codeberg.org" in url:
        return "codeberg"
    if "gitlab.com" in url:
        return "gitlab"
    if "bitbucket.org" in url:
        return "bitbucket"
    if "sr.ht" in url:
        return "sourcehut"
    if "forgejo" in url:
        return "forgejo"
    if "gitea.com" in url or "gitea." in url:
        return "gitea"
    if "notabug.org" in url:
        return "notabug"
    return "remote"


@dataclass
class Info:
    path: str = ""
    branch: str = "unknown"
    status: Status = field(default_factory=Status)
    health: Health = Health.UNKNOWN
    ahead: int = 0
    behind: int = 0
    last_commit: Optional[Commit] = None
    remotes: list[Remote] = field(default_factory=list)
    detached: bool = False
    error: Optional[str] = None

    def has_remote(self) -> bool:
        return len(self.remotes) > 0

    def remote_summary(self) -> str:
        if not self.remotes:
            return "local"
        labels: list[str] = []
        seen: set[str] = set()
        for r in self.remotes:
            label = _remote_label(r.url)
            if label not in seen:
                seen.add(label)
                labels.append(label)
        return "+".join(labels)


def _run_git(path: str, *args: str) -> str:
    try:
        result = subprocess.run(
            ["git", "-C", path, *args],
            capture_output=True,
            text=True,
            timeout=10,
        )
        if result.returncode != 0:
            return ""
        return result.stdout.strip()
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return ""


def get_info(path: str) -> Info:
    info = Info(path=path)

    branch = _get_branch(path)
    if not branch:
        info.error = "could not get branch"
        return info
    info.branch = branch
    info.detached = branch == "HEAD"

    status = _get_status(path)
    if status:
        info.status = status

    commit = _get_last_commit(path)
    if commit:
        info.last_commit = commit

    ahead, behind = _get_ahead_behind(path)
    info.ahead = ahead
    info.behind = behind

    remotes = _get_remotes(path)
    info.remotes = remotes

    info.health = _determine_health(info)
    return info


def _get_branch(path: str) -> str:
    return _run_git(path, "rev-parse", "--abbrev-ref", "HEAD")


def _get_status(path: str) -> Optional[Status]:
    output = _run_git(path, "status", "--porcelain")
    if output == "":
        return Status(is_clean=True)

    s = Status(is_clean=True)
    for line in output.split("\n"):
        if len(line) < 2:
            continue
        x, y = line[0], line[1]
        if x == "?" and y == "?":
            s.untracked += 1
        else:
            if x != " " and x != "?":
                s.staged += 1
            if y != " " and y != "?":
                s.modified += 1
            if y == "?":
                s.untracked += 1

    s.is_clean = s.modified == 0 and s.untracked == 0 and s.staged == 0
    return s


def _get_last_commit(path: str) -> Optional[Commit]:
    output = _run_git(path, "log", "-1", "--format=%s|%ct")
    if not output:
        return None

    parts = output.split("|", 1)
    if len(parts) != 2:
        return None

    try:
        timestamp = float(parts[1])
    except ValueError:
        return None

    return Commit(message=parts[0], date=timestamp)


def _get_remotes(path: str) -> list[Remote]:
    output = _run_git(path, "remote", "-v")
    if not output:
        return []

    seen: set[str] = set()
    remotes: list[Remote] = []
    for line in output.split("\n"):
        parts = line.split()
        if len(parts) < 2:
            continue
        name, url = parts[0], parts[1]
        key = f"{name}|{url}"
        if key not in seen:
            seen.add(key)
            remotes.append(Remote(name=name, url=url))
    return remotes


def _get_ahead_behind(path: str) -> tuple[int, int]:
    output = _run_git(path, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
    if not output:
        return 0, 0

    parts = output.split()
    if len(parts) != 2:
        return 0, 0

    try:
        return int(parts[0]), int(parts[1])
    except ValueError:
        return 0, 0


def _determine_health(info: Info) -> Health:
    if info.detached:
        return Health.DETACHED
    if not info.status.is_clean:
        return Health.DIRTY
    if info.ahead > 0 and info.behind > 0:
        return Health.DIVERGED
    if info.ahead > 0:
        return Health.AHEAD
    if info.behind > 0:
        return Health.BEHIND
    return Health.CLEAN
