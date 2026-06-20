from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

from delta.git_info import Info, get_info

_SKIP_DIRS = frozenset({
    "node_modules", ".venv", "venv", "__pycache__",
    ".next", ".nuxt", "dist", "build", ".cache",
    "target", ".gradle", ".idea", ".vscode",
    "env", ".env", "vendor", "Pods", "bin", "obj",
    ".git", ".hg", ".svn", "site-packages",
    ".langchain-venv", "ephemeral", "Runner",
    "RunnerTests", "Flutter", "__pypackages__",
})

_PROJECT_MARKERS = frozenset({
    "package.json", "go.mod", "Cargo.toml",
    "requirements.txt", "pyproject.toml",
    "Makefile", "CMakeLists.txt", "pom.xml",
    "build.gradle", "package-lock.json",
    "yarn.lock", "pnpm-lock.yaml",
    "tsconfig.json", "composer.json",
    "setup.py", "setup.cfg",
})

_CODE_EXTS = frozenset({
    ".go", ".py", ".js", ".ts", ".jsx", ".tsx",
    ".c", ".cpp", ".h", ".hpp", ".rs", ".java",
    ".rb", ".php", ".cs", ".swift", ".kt",
})


@dataclass
class Repo:
    path: str
    name: str
    is_git: bool = False
    git_info: Optional[Info] = field(default=None)


def scan_folders(folders: list[str]) -> list[Repo]:
    repos: list[Repo] = []
    for folder in folders:
        folder = folder.strip().strip('"')
        folder = os.path.normpath(folder)

        if not os.path.isdir(folder):
            continue

        _walk_and_collect(folder, repos, inside_git=False)

    return _dedup(repos)


def scan_folders_with_git_info(folders: list[str]) -> list[Repo]:
    repos = scan_folders(folders)
    for repo in repos:
        if repo.is_git:
            repo.git_info = get_info(repo.path)
    return repos


def _walk_and_collect(root: str, repos: list[Repo], inside_git: bool) -> None:
    try:
        entries = os.listdir(root)
    except PermissionError:
        return

    has_git = ".git" in entries

    if has_git:
        repos.append(Repo(
            path=root,
            name=os.path.basename(root),
            is_git=True,
        ))
        for entry in entries:
            full = os.path.join(root, entry)
            if not os.path.isdir(full):
                continue
            if _should_skip_dir(entry):
                continue
            _walk_and_collect(full, repos, inside_git=True)
        return

    if not inside_git and _is_non_git_code_folder(root) and not _has_sub_repo(root):
        repos.append(Repo(
            path=root,
            name=os.path.basename(root),
            is_git=False,
        ))
        return

    for entry in entries:
        full = os.path.join(root, entry)
        if not os.path.isdir(full):
            continue
        if _should_skip_dir(entry):
            continue
        _walk_and_collect(full, repos, inside_git=inside_git)


def _should_skip_dir(name: str) -> bool:
    if name in _SKIP_DIRS:
        return True
    if name.startswith("."):
        return True
    return False


def _has_sub_repo(path: str) -> bool:
    try:
        entries = os.listdir(path)
    except PermissionError:
        return False
    for entry in entries:
        if not os.path.isdir(os.path.join(path, entry)):
            continue
        if _should_skip_dir(entry):
            continue
        if os.path.isdir(os.path.join(path, entry, ".git")):
            return True
    return False


def _is_non_git_code_folder(path: str) -> bool:
    try:
        entries = os.listdir(path)
    except PermissionError:
        return False

    for entry in entries:
        if entry == ".git":
            return False
        if entry in _PROJECT_MARKERS:
            return True
        _, ext = os.path.splitext(entry)
        if ext.lower() in _CODE_EXTS:
            return True
    return False


def _dedup(repos: list[Repo]) -> list[Repo]:
    seen: set[str] = set()
    result: list[Repo] = []
    for repo in repos:
        if repo.path not in seen:
            seen.add(repo.path)
            result.append(repo)
    return result
