from dataclasses import dataclass, field
import yaml
import os
import re
import subprocess
from typing import List, Dict, Union

import tasks.context


@dataclass
class Settings:
    archive_dir: str = '~'
    clone_dir: str = '~'


@dataclass
class UnlessCmd:
    cmd: str
    post: str = None

    def get_fn(self, operation: str, parameter: int):
        if operation == 'head':
            return lambda x: x.split('\n')[parameter]
        elif operation == 'split':
            return lambda x: x.split()[parameter]
        else:
            raise Exception(f'Unknown operation {operation}')

    def get_version(self, output, version_fn):
        ops = []

        for op in version_fn.split('|'):
            operation, parameter = op.strip().split()
            parameter = int(parameter)
            ops.append(self.get_fn(operation, parameter))

        for op in ops:
            output = op(output)

        return output

    def unless(self, version: str = ''):
        proc = subprocess.run(self.cmd, shell=True, capture_output=True, text=True)
        if proc.returncode != 0:
            return True

        if not self.post:
            return False

        if not version:
            return

        output = proc.stdout.strip()
        current_version = self.get_version(output, self.post)
        if current_version == version:
            return False

        return True


@dataclass
class UnlessFile:
    ls: str

    def unless(self, context: Dict = None):
        if not context:
            context = {}
        ls_target = tasks.context.expand(self.ls, context)
        ls_target = os.path.expanduser(ls_target)
        return not os.path.exists(ls_target)


@dataclass
class Archive:
    url: str
    unless: Union[UnlessCmd, UnlessFile] = None
    version: str = ''
    symlink: str = ''


@dataclass
class CargoCrate:
    name: str
    unless: Union[UnlessCmd, UnlessFile] = None
    bins: bool = False


@dataclass
class GoPkg:
    name: str
    unless: Union[UnlessCmd, UnlessFile] = None
    host: str = 'github.com'
    version: str = 'latest'


@dataclass
class Template:
    src: str
    dest: str
    context: Dict[str, str] = field(default_factory=dict)
    mode: str = '0644'


@dataclass
class Service:
    name: str
    user: bool = True
    start: bool = True
    enable: bool = True
    description: str = ''
    exec_: str = ''
    env: List[Dict[str, str]] = field(default_factory=list)


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
