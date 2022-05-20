import os
from typing import List

from pyinfra.api import FunctionCommand, operation

from tasks.recipes import run_command
#from tasks.github_keys import write_lines

KNOWN_HOSTS = os.path.expanduser('~/.ssh/known_hosts')


def get_host_keys(host):
    out = run_command(f'ssh-keyscan {host}')
    return out.split('\n')


def ensure_keys(host):
    host_keys = get_host_keys(host)
    with open(KNOWN_HOSTS, 'a') as f:
        for key in host_keys:
            f.write(f'{key}\n')


@operation
def ensure_host_keys(hosts: List[str]):
    missing_hosts = set()

    for host in hosts:
        output = run_command(f'ssh-keygen -F {host}', raise_on_error=False)
        if output:
            continue
        missing_hosts.add(host)

    for host in missing_hosts:
        yield FunctionCommand(ensure_keys, [host], {})


def run(config):
    ensure_host_keys(config.accept_host_keys)
