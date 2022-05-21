from dataclasses import dataclass, field
import os
from typing import Dict, List, Union

from pyinfra.api import FunctionCommand, operation
from pyinfra import host

import facts.base
import tasks.archives
import tasks.config
from tasks.context import expand
from tasks.ops import run_command
from tasks.templates import do_template_file, maybe_template_file
from tasks.unless import UnlessCmd, UnlessFile


@dataclass
class Recipe:
    task: str
    unless: Union[UnlessCmd, UnlessFile] = None
    steps: List[Dict] = field(default_factory=list)
    when: str = ''


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


@dataclass
class Template:
    name: str
    target: str
    mode: str = '0644'
    content: str = ''

    def run(self):
        target = expand(self.target)
        template = tasks.config.Template(dest=target, mode=self.mode)
        if update_op := maybe_template_file(template, output=self.content):
            do_template_file(update_op)


STEP_CLASSES = {
    'cmd': Cmd,
    'download': Download,
    'git': Git,
    'quicklisp': Quicklisp,
    'rename': Rename,
    'symlink': Symlink,
    'template': Template,
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
