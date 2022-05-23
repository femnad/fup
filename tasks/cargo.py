from pyinfra.api import StringCommand, operation

from tasks.archives import get_unless
from tasks.config import CargoCrate


@operation
def install_crate(crate: CargoCrate):
    if (unless := get_unless(crate.unless)) and not unless.should_proceed():
        return

    maybe_bins = ' --bins' if crate.bins else ''
    maybe_git = '--git ' if crate.name.startswith('https://git') else ''

    yield StringCommand(f'cargo install {maybe_git}{crate.name}{maybe_bins}')


def run(config):
    for crate in config.cargo:
        crate = CargoCrate(**crate)
        install_crate(crate)
