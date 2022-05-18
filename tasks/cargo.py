from pyinfra.api import StringCommand, operation

from tasks.archives import get_unless
from tasks.config import CargoCrate


@operation
def install_crate(crate: CargoCrate):
    unless = get_unless(crate.unless)
    if not unless.unless():
        return
    yield StringCommand(f'cargo install {crate.name}')


def run(config):
    for crate in config.cargo:
        crate = CargoCrate(**crate)
        install_crate(crate)
