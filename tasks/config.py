from dataclasses import dataclass
import yaml
import os
import re
from typing import List, Dict, Union


@dataclass
class Settings:
    archive_dir: str
    clone_dir: str


@dataclass
class UnlessCmd:
    cmd: str
    post: str = None


@dataclass
class UnlessFile:
    ls: str


@dataclass
class Archive:
    url: str
    unless: Union[UnlessCmd, UnlessFile] = None
    version: str = ''
    symlink: str = ''


@dataclass
class Config:
    packages: Dict[str, List[str]]
    archives: List[Archive]
    settings: Settings
    recipes: Dict


def get_config():
    with open(os.path.expanduser('~/.config/fup/fup.yml')) as f:
        cfg_dict = yaml.load(f, Loader=yaml.SafeLoader)
        cfg = Config(**cfg_dict)
        cfg.settings = Settings(**cfg_dict['settings'])
        cfg.archives = [Archive(**a) for a in cfg.archives]
        return cfg


def get_packages(cfg, current_dist_id):
    package_set = set()
    for dist_id, packages in cfg.packages.items():
        is_re = re.escape(dist_id) != dist_id
        if is_re and (_ := re.match(dist_id, current_dist_id)):
            package_set.update(packages)
        elif dist_id == current_dist_id:
            package_set.update(packages)

    return sorted(package_set)
