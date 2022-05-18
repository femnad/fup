from dataclasses import dataclass, field
import os
import subprocess
from typing import Dict, List, Union

from pyinfra.api import FunctionCommand, operation
from tasks.config import UnlessCmd, UnlessFile
import tasks.archives


@dataclass
class Recipe:
    task: str
    unless: Union[UnlessCmd, UnlessFile]
    steps: List[Dict]


def run_command(cmd, pwd=None, sudo=False):
    prev_dir = None

    if sudo:
        cmd = f'sudo {cmd}'

    if pwd:
        pwd = os.path.expanduser(pwd)
        prev_dir = os.getcwd()
        os.chdir(pwd)

    proc = subprocess.run(cmd, shell=True, text=True, capture_output=True)

    if proc.returncode == 0:
        if prev_dir:
            os.chdir(prev_dir)
        return

    output = {}
    if proc.stdout:
        output['stdout'] = proc.stdout.strip()
    if proc.stderr:
        output['stderr'] = proc.stderr.strip()
    msg = '\n'.join([f'{k}: {v}' for k, v in output.items()])
    raise Exception(f'Error running command {cmd}\n{msg}')


@dataclass
class Download:
    name: str
    url: str
    target: str
    unless: Union[UnlessCmd, UnlessFile] = None

    def run(self):
        target = os.path.expanduser(self.target)
        tasks.archives.download(self.url, target)


@dataclass
class Cmd:
    name: str
    cmd: str
    pwd: str = ''
    sudo: bool = False

    def sh(self, cmd):
        run_command(cmd, pwd=self.pwd, sudo=self.sudo)

    def run(self):
        if '\n' not in self.cmd:
            self.sh(self.cmd)
            return

        for cmd in self.cmd.split('\n'):
            self.sh(cmd)


@dataclass
class Rename:
    name: str
    src: str
    dst: str

    def run(self):
        src = os.path.expanduser(self.src)
        dst = os.path.expanduser(self.dst)
        if not os.path.exists(src):
            return
        os.rename(src, dst)


@dataclass
class Quicklisp:
    name: str
    pkg: list = field(default_factory=list)

    def run(self):
        for p in self.pkg:
            run_command(f"sbcl --eval '(ql:quickload \"{p}\")' --non-interactive")


@dataclass
class Git:
    name: str
    repo: str
    target: str

    def run(self):
        target = os.path.expanduser(self.target)
        if os.path.exists(target):
            return
        run_command(f'git clone {self.repo} {target}')


STEP_CLASSES = {Download, Cmd, Rename, Quicklisp, Git}


def try_get_step(step, cls):
    try:
        return cls(**step)
    except TypeError:
        return


def get_step_class(content, settings):
    for cls in STEP_CLASSES:
        content = {
            k: tasks.context.expand(v, settings.__dict__) if isinstance(v, str) else v
            for k, v in content.items()
        }
        if step := try_get_step(content, cls):
            return step

    raise Exception(f'Cannot determine step type for {content}')


def do_run_recipe(steps, settings):
    for step in steps:
        s = get_step_class(step, settings)
        s.run()


@operation
def run_recipe(recipe, settings):
    unless = tasks.archives.get_unless(recipe.unless)
    if not unless.unless():
        return
    yield FunctionCommand(do_run_recipe, [recipe.steps, settings], {})


def run(config):
    for recipe in config.recipes:
        recipe = Recipe(**recipe)
        run_recipe(recipe, config.settings)
