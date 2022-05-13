from dataclasses import dataclass
import yaml
from typing import List, Dict, Union
import os

BASE_PACKAGES_KEY = 'base'


@dataclass
class Settings:
    archive_dir: str


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


def get_config():
    with open(os.path.expanduser('~/.config/fup/fup.yml')) as f:
        cfg_dict = yaml.load(f, Loader=yaml.SafeLoader)
        cfg = Config(**cfg_dict)
        cfg.settings = Settings(**cfg_dict['settings'])
        cfg.archives = [Archive(**a) for a in cfg.archives]
        return cfg


def get_packages(cfg, dist_id):
    package_set = set(cfg.packages.get(BASE_PACKAGES_KEY, []))

    for oss, packages in cfg.packages.items():
        oss = oss.split('|')
        if dist_id in oss:
            package_set.update(packages)

    return package_set
