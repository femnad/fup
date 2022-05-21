import os

from pyinfra.operations import files

from tasks.config import EnsureLine


def replace(ensure_line: EnsureLine, should_sudo: bool):
    files.replace(path=ensure_line.file, text=ensure_line.text, replace=ensure_line.replace, _sudo=should_sudo)


ENSURE_FNS = {
    'replace': replace,
}


def ensure(ensure_line: EnsureLine, should_sudo: bool):
    name = ensure_line.name

    if name not in ENSURE_FNS:
        raise Exception(f'Cannot ensure line with operation {name}')

    fn = ENSURE_FNS[name]
    fn(ensure_line, should_sudo)


def run(config):
    user_home = os.getenv('HOME')

    for line in config.ensure_lines:
        line = EnsureLine(**line)
        should_sudo = not line.file.startswith(user_home)
        ensure(line, should_sudo)
