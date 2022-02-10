from dataclasses import dataclass
import yaml
from typing import List, Dict
import os


@dataclass
class Config:
    packages: Dict[str, List[str]]


def get_packages(dist_id):
    package_set = set()

    with open(os.path.expanduser('~/.config/fup/fup.yml')) as f:
        cfg = Config(**yaml.load(f, Loader=yaml.SafeLoader))
        for oss, packages in cfg.packages.items():
            oss = oss.split('|')
            if dist_id in oss:
                package_set.update(packages)

    return package_set
