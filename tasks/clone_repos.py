import os
from typing import List

from pyinfra import host
from pyinfra.operations import git
from pyinfra.api import operation, FunctionCommand

import facts.base
import tasks.config
import tasks.ops
import tasks.when


def is_self_repo(repo: tasks.config.Repo, settings: tasks.config.Settings) -> bool:
    repo_owner, _ = repo.name.split('/')[-2:]
    return repo_owner == settings.github_user


def is_self_repo_ready():
    return host.get_fact(facts.base.SshReady)


def get_clone_dir(repo: tasks.config.Repo, settings: tasks.config.Settings):
    repo_owner, repo_name = repo.name.split('/')[-2:]
    base_dir = settings.self_clone_dir if repo_owner == settings.github_user else settings.clone_dir

    return os.path.expanduser(f'{base_dir}/{repo_name}')


def get_clone_url(repo: tasks.config.Repo, settings: tasks.config.Settings):
    repo_owner, _ = repo.name.split('/')[-2:]

    if repo_owner == settings.github_user:
        return f'git@{repo.host}:{repo.name}.git'

    return f'https://{repo.host}/{repo.name}.git'


def maybe_add_remote(path, remote, url):
    if tasks.ops.run_command(f'git remote get-url {remote}', pwd=path, raise_on_error=False):
        return

    tasks.ops.run_command(f'git remote add {remote} {url}', pwd=path)


def checkout_ref(path, ref):
    tasks.ops.run_command(f'git checkout {ref}', pwd=path)


@operation
def clone_repos(repos: List[tasks.config.Repo], settings: tasks.config.Settings):
    for repo in repos:
        if not tasks.when.should_run(repo.when):
            continue

        if is_self_repo(repo, settings) and not is_self_repo_ready():
            continue

        clone_dir = get_clone_dir(repo, settings)
        url = get_clone_url(repo, settings)

        yield from git.repo(src=url,
                            dest=clone_dir,
                            update_submodules=repo.submodule,
                            recursive_submodules=repo.recursive_submodule)

        if repo.ref:
            yield FunctionCommand(checkout_ref, [clone_dir, repo.ref], {})

        for remote, url in repo.remotes.items():
            yield FunctionCommand(maybe_add_remote, [clone_dir, remote, url], {})


def run(config):
    repos = [tasks.config.Repo(**repo) for repo in config.repos]
    self_repos = [tasks.config.Repo(**repo) for repo in config.repos]
    clone_repos(repos, config.settings)
    clone_repos(self_repos, config.settings)
