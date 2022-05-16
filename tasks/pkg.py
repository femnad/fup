from pyinfra.operations import apt, dnf
from pyinfra import host
from pyinfra.facts.server import LinuxDistribution

import tasks.config

INSTALLERS = {
    apt: {'debian', 'ubuntu'},
    dnf: {'fedora'},
}


def get_installer(dist_id):
    id_to_installer = {}
    for installer, ids in INSTALLERS.items():
        for os_id in ids:
            id_to_installer[os_id] = installer

    return id_to_installer[dist_id]


def install(cfg: tasks.config.Config):
    dist_id = host.get_fact(LinuxDistribution)['release_meta']['ID']
    pkgs = tasks.config.get_packages(cfg, dist_id)

    installer = get_installer(dist_id)

    installer.packages(packages=pkgs, _sudo=True)
