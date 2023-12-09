import os
import shlex
import subprocess
import yaml


def run_fup(config: dict, config_file: str, provisioner: str) -> int:
    config_dir = os.path.dirname(config_file)
    if not os.path.exists(config_dir):
        os.mkdir(config_dir)

    with open(config_file, 'wt') as fd:
        yaml.dump(config, Dumper=yaml.SafeDumper, stream=fd)

    go_path = os.getenv('GOPATH', os.path.expanduser('~/go'))
    fup_bin = os.path.join(go_path, 'bin', 'fup')
    cmd = shlex.split(f'{fup_bin} -p {provisioner} -f {config_file} -l 0 -n')
    return subprocess.run(cmd).returncode
