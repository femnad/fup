from dataclasses import dataclass, field
import yaml
import os
import re
from typing import List, Dict, Union

import tasks.unless


@dataclass
class Settings:
    archive_dir: str = '~'
    clone_dir: str = '~'
    host_facts: Dict[str, Dict[str, str]] = field(default_factory=dict)


@dataclass
class Archive:
    url: str
    unless: Union[tasks.unless.UnlessCmd, tasks.unless.UnlessFile] = None
    version: str = ''
    symlink: str = ''


@dataclass
class CargoCrate:
    name: str
    unless: Union[tasks.unless.UnlessCmd, tasks.unless.UnlessFile] = None
    bins: bool = False


@dataclass
class GoPkg:
    name: str
    unless: Union[tasks.unless.UnlessCmd, tasks.unless.UnlessFile] = None
    host: str = 'github.com'
    version: str = 'latest'


@dataclass
class Template:
    src: str
    dest: str
    mode: str = '0644'
    context: Dict[str, str] = field(default_factory=dict)


@dataclass
class ServiceUnit:
    exec: str
    description: str
    before: str = ''
    type: str = ''
    wanted_by: str = 'default'
    environment: List[Dict[str, str]] = field(default_factory=list)
    service_options: Dict[str, str] = field(default_factory=dict)


@dataclass
class Service:
    name: str
    user: bool = True
    start: bool = True
    enable: bool = True
    unit: ServiceUnit = None
    context: Dict[str, str] = field(default_factory=dict)


@dataclass
class GithubUserKeys:
    user: str = ''


@dataclass
class Repo:
    name: str
    host: str = 'github.com'


@dataclass
class Config:
    packages: Dict[str, List[str]] = field(default_factory=dict)
    archives: List[Archive] = field(default_factory=list)
    settings: Settings = Settings()
    recipes: Dict = field(default_factory=dict)
    cargo: List[CargoCrate] = field(default_factory=list)
    gopkg: List[GoPkg] = field(default_factory=list)
    templates: List[Template] = field(default_factory=list)
    services: List[Service] = field(default_factory=list)
    github_user_keys: GithubUserKeys = field(default_factory=dict)
    accept_host_keys: List[str] = field(default_factory=list)
    repos: List[Repo] = field(default_factory=list)


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
