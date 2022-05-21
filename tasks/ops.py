import os
import subprocess
from typing import List


def run_commands(cmds: List[str], pwd=None, sudo=False, raise_on_error=True, env=None):
    for cmd in cmds:
        run_command(cmd, pwd, sudo, raise_on_error, env)


def run_command(cmd, pwd=None, sudo=False, raise_on_error=True, env=None):
    prev_dir = None

    if sudo:
        cmd = f'sudo {cmd}'

    if pwd:
        pwd = os.path.expanduser(pwd)
        prev_dir = os.getcwd()
        os.chdir(pwd)

    proc = subprocess.run(cmd, shell=True, text=True, capture_output=True, env=env)

    if proc.returncode == 0 or not raise_on_error:
        if prev_dir:
            os.chdir(prev_dir)
        return proc.stdout.strip()

    output = {}
    if proc.stdout:
        output['stdout'] = proc.stdout.strip()
    if proc.stderr:
        output['stderr'] = proc.stderr.strip()
    msg = '\n'.join([f'{k}: {v}' for k, v in output.items()])

    raise Exception(f'Error running command {cmd}\n{msg}')
