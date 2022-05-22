import re

from pyinfra.operations import apt, dnf
from pyinfra import host
from pyinfra.facts.server import LinuxDistribution

import tasks.config

INSTALLERS = {
    apt: {'debian', 'ubuntu'},
    dnf: {'fedora'},
}


def get_packages(current_dist_id, packages):
    package_set = set()
    for dist_id, packages in packages.items():
        is_re = re.escape(dist_id) != dist_id
        if is_re and (_ := re.match(dist_id, current_dist_id)):
            package_set.update(packages)
        elif dist_id == current_dist_id:
            package_set.update(packages)

    return sorted(package_set)


def get_installer(dist_id):
    id_to_installer = {}
    for installer, ids in INSTALLERS.items():
        for os_id in ids:
            id_to_installer[os_id] = installer

    return id_to_installer[dist_id]


def run(cfg: tasks.config.Config):
    dist_id = host.get_fact(LinuxDistribution)['release_meta']['ID']

    pkgs = get_packages(dist_id, cfg.packages)
    unwanted_pkgs = get_packages(dist_id, cfg.unwanted_packages)

    installer = get_installer(dist_id)

    installer.packages(packages=pkgs, _sudo=True)
    installer.packages(packages=unwanted_pkgs, present=False, _sudo=True)
