from dataclasses import dataclass
import yaml
from typing import List, Dict
import os


@dataclass
class Settings:
    archive_dir: str
    vars: Dict[str, str]


@dataclass
class Config:
    packages: Dict[str, List[str]]
    archives: List[Dict[str, str]]
    settings: Settings


def get_config():
    with open(os.path.expanduser('~/.config/fup/fup.yml')) as f:
        cfg_dict = yaml.load(f, Loader=yaml.SafeLoader)
        cfg = Config(**cfg_dict)
        cfg.settings = Settings(**cfg_dict['settings'])
        return cfg


def get_packages(cfg, dist_id):
    package_set = set()

    for oss, packages in cfg.packages.items():
        oss = oss.split('|')
        if dist_id in oss:
            package_set.update(packages)

    return package_set
