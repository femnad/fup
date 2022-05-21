from typing import List

from pyinfra.api import StringCommand, operation

from tasks.archives import get_unless
from tasks.config import GoPkg


def install_pkg(go_pkg: GoPkg):
    pkg = f'{go_pkg.host}/{go_pkg.name}@{go_pkg.version}'
    yield StringCommand(f'go install {pkg}')


@operation
def install_pkgs(go_pkgs: List[GoPkg]):
    for pkg in go_pkgs:
        unless = get_unless(pkg.unless)
        if not unless.should_proceed():
            continue

        yield from install_pkg(pkg)


def run(config):
    pkgs = [GoPkg(**pkg) for pkg in config.gopkg]
    install_pkgs(pkgs)
