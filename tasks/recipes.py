from dataclasses import dataclass, field
import os
import subprocess
from typing import Dict, List, Union

from pyinfra.api import FunctionCommand, operation
from pyinfra import host

import tasks.archives
import facts.base
from tasks.context import expand
from tasks.unless import UnlessCmd, UnlessFile


@dataclass
class Recipe:
    task: str
    unless: Union[UnlessCmd, UnlessFile] = None
    steps: List[Dict] = field(default_factory=list)
    when: str = ''


def run_command(cmd, pwd=None, sudo=False, raise_on_error=True):
    prev_dir = None

    if sudo:
        cmd = f'sudo {cmd}'

    if pwd:
        pwd = os.path.expanduser(pwd)
        prev_dir = os.getcwd()
        os.chdir(pwd)

    proc = subprocess.run(cmd, shell=True, text=True, capture_output=True)

    if proc.returncode == 0 or not raise_on_error:
        if prev_dir:
            os.chdir(prev_dir)
        return proc.stdout.strip()

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
    unless: Union[UnlessCmd, UnlessFile] = None

    def sh(self, cmd):
        run_command(cmd, pwd=self.pwd, sudo=self.sudo)

    def run(self):
        if (unless := tasks.archives.get_unless(self.unless)) and not unless.should_proceed():
            return

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


@dataclass
class Symlink:
    name: str
    link: str
    target: str

    def run(self):
        link = expand(self.link)
        target = expand(self.target)
        # use `lexists` to detect broken links
        if os.path.lexists(link):
            if os.path.realpath(link) == target:
                return
            else:
                os.unlink(link)
        os.symlink(src=target, dst=link)


STEP_CLASSES = {
    'cmd': Cmd,
    'download': Download,
    'git': Git,
    'quicklisp': Quicklisp,
    'rename': Rename,
    'symlink': Symlink,
}


def get_step_class(step_content, settings):
    step_name = step_content['name']

    if step_name not in STEP_CLASSES:
        raise Exception(f'Cannot find step class for {step_name}')

    step_class = STEP_CLASSES[step_name]
    content = {
        k: tasks.context.expand(v, settings.__dict__) if isinstance(v, str) else v
        for k, v in step_content.items()
    }

    return step_class(**content)


def do_run_recipe(steps, settings):
    for step in steps:
        s = get_step_class(step, settings)
        s.run()


def should_run(when):
    if not when:
        return True

    negate = False
    if ' ' in when and when.startswith('not '):
        negate = True
        when = when.split()[-1]

    fact_class = ''.join([w.capitalize() for w in when.split('-')])
    result = host.get_fact(getattr(facts.base, fact_class))

    return not result if negate else result


@operation
def run_recipe(recipe, settings):
    if not should_run(recipe.when):
        return

    unless = tasks.archives.get_unless(recipe.unless)
    if unless and not unless.should_proceed():
        return

    yield FunctionCommand(do_run_recipe, [recipe.steps, settings], {})


def run(config):
    for recipe in config.recipes:
        recipe = Recipe(**recipe)
        run_recipe(recipe, config.settings)
