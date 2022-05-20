import os
from typing import List

from tasks.config import Repo
from tasks.recipes import run_command


from pyinfra.api import StringCommand, operation


@operation
def clone_repos(repos: List[Repo], clone_dir):
    for repo in repos:
        name = os.path.basename(repo.name)
        path = os.path.expanduser(f'{clone_dir}/{name}')
        if os.path.exists(path):
            continue
        url = f'https://{repo.host}/{repo.name}.git'
        yield StringCommand(f'git clone {url} {path}')


def run(config):
    repos = [Repo(**repo) for repo in config.repos]
    clone_repos(repos, config.settings.clone_dir)
