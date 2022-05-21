import os
from typing import List

from pyinfra.api import FunctionCommand, operation

from tasks.config import PipPkg
from tasks.context import expand
from tasks.ops import run_command


def install_package(pkg, virtualenv_base):
    virtualenv = f"{virtualenv_base}/{pkg.name}"

    if not os.path.exists(virtualenv):
        run_command(f"virtualenv {virtualenv}")

    pip = f"{virtualenv}/bin/pip"

    run_command(f"{pip} install {pkg.name}")

    for req in pkg.reqs:
        run_command(f"{pip} install {req}")


@operation
def install_packages(pkgs: List[PipPkg], virtualenv_path: str):
    virtualenv_path = expand(virtualenv_path)
    for pkg in pkgs:
        yield FunctionCommand(install_package, [pkg, virtualenv_path], {})


def run(config):
    pip_pkgs = [PipPkg(**p) for p in config.pip_pkgs]
    install_packages(pip_pkgs, config.settings.virtualenv_dir)
