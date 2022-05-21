import os
from typing import List

from tasks.config import Repo, Settings
import tasks.when

from pyinfra.operations import git
from pyinfra.api import operation


def get_clone_dir(repo: Repo, settings: Settings):
    repo_owner, repo_name = repo.name.split('/')[-2:]
    base_dir = settings.self_clone_dir if repo_owner == settings.github_user else settings.clone_dir

    return os.path.expanduser(f'{base_dir}/{repo_name}')


def get_clone_url(repo: Repo, settings: Settings):
    repo_owner, _ = repo.name.split('/')[-2:]

    if repo_owner == settings.github_user:
        return f'git@{repo.host}:{repo.name}.git'

    return f'https://{repo.host}/{repo.name}.git'


@operation
def clone_repos(repos: List[Repo], settings: Settings):
    for repo in repos:
        if not tasks.when.should_run(repo.when):
            continue

        clone_dir = get_clone_dir(repo, settings)
        url = get_clone_url(repo, settings)

        yield from git.repo(src=url, dest=clone_dir)


def run(config):
    repos = [Repo(**repo) for repo in config.repos]
    clone_repos(repos, config.settings)
