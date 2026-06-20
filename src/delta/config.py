from __future__ import annotations

import os
from pathlib import Path
from typing import Optional

import yaml
from pydantic import BaseModel, Field


class ColumnsConfig(BaseModel):
    repo: bool = True
    branch: bool = True
    status: bool = True
    health: bool = True
    remote: bool = True
    last_commit: bool = True
    path: str = "short"  # full | short | hidden


class StaleConfig(BaseModel):
    enabled: bool = True
    threshold_days: int = 90


class Config(BaseModel):
    scan_folders: list[str] = Field(default_factory=list)
    editor: str = "code"
    columns: ColumnsConfig = Field(default_factory=ColumnsConfig)
    stale: StaleConfig = Field(default_factory=StaleConfig)


def default_path() -> str:
    home = os.path.expanduser("~")
    return os.path.join(home, ".config", "delta", "config.yaml")


def load(path: str = "") -> Config:
    if not path:
        path = default_path()

    if not os.path.exists(path):
        cfg = Config()
        save(cfg, path)
        return cfg

    with open(path, "r", encoding="utf-8") as f:
        data = yaml.safe_load(f)

    if data is None:
        return Config()

    return Config(**data)


def save(cfg: Config, path: str = "") -> None:
    if not path:
        path = default_path()

    os.makedirs(os.path.dirname(path), exist_ok=True)

    data = cfg.model_dump()
    with open(path, "w", encoding="utf-8") as f:
        yaml.dump(data, f, default_flow_style=False, sort_keys=False)


def add_folder(cfg: Config, folder: str) -> None:
    folder = folder.strip().strip('"')
    folder = os.path.normpath(folder)
    folder = os.path.expanduser(folder)
    if folder in cfg.scan_folders:
        raise ValueError(f"folder already in config: {folder}")
    cfg.scan_folders.append(folder)
