import json
import os

from tasks.archives import http_request
from tasks.config import GithubUserKeys

from pyinfra.api import FunctionCommand, operation

AUTHORIZED_KEYS_FILE = os.path.expanduser('~/.ssh/authorized_keys')


def write_lines(lines):
    with open(AUTHORIZED_KEYS_FILE, 'a') as f:
        for line in lines:
            f.write(f'{line}\n')


@operation
def ensure_github_user_keys(cfg: GithubUserKeys):
    if not cfg.user:
        return

    url = f'https://api.github.com/users/{cfg.user}/keys'
    keys = http_request(url, 'GET', output_file=None)
    keys = json.loads(keys)
    missing_keys = set([key['key'] for key in keys])

    with open(AUTHORIZED_KEYS_FILE, 'r') as f:
        for line in f:
            line = line.strip()
            if line in missing_keys:
                missing_keys.remove(line)

    if not missing_keys:
        return

    yield FunctionCommand(write_lines, [missing_keys], {})


def run(config):
    keys = GithubUserKeys(**config.github_user_keys)
    ensure_github_user_keys(keys)
