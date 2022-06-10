from typing import List

from pyinfra.api import StringCommand, operation

from tasks.archives import get_unless
from tasks.config import GoPkg


def get_pkg_name(go_pkg: GoPkg) -> str:
    name = go_pkg.name
    fields = name.split('/')

    if '.' in fields[0]:
        return name

    return f'{go_pkg.host}/{name}'


def install_pkg(go_pkg: GoPkg):
    name = get_pkg_name(go_pkg)
    pkg = f'{name}@{go_pkg.version}'
    yield StringCommand(f'go install {pkg}')


@operation
def install_pkgs(go_pkgs: List[GoPkg]):
    for pkg in go_pkgs:
        if (unless := get_unless(pkg.unless)) and not unless.should_proceed():
            continue

        yield from install_pkg(pkg)


def run(config):
    pkgs = [GoPkg(**pkg) for pkg in config.gopkg]
    install_pkgs(pkgs)
