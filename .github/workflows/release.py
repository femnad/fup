#!/usr/bin/env python3
import os
import re
import subprocess
import shlex
import sys

MAIN_FILE = 'main.go'
VERSION_LINE = re.compile('version = "([0-9]+.[0-9]+.[0-9]+)"')


def sh(cmd: str, env: dict | None = None) -> str:
    cmd_parsed = shlex.split(cmd)
    proc = subprocess.run(cmd_parsed, text=True, capture_output=True, env=env)
    stdout = proc.stdout.strip()
    code = proc.returncode
    if code:
        out = ''
        stderr = proc.stderr.strip()
        if stderr:
            out += f'stderr: {stderr}'
        if stdout:
            prefix = ', ' if out else ''
            out += f'{prefix}stdout: {stdout}'
        print(f'Command `{cmd}` exited with code {code}, output: {out}')
        sys.exit(code)

    return proc.stdout.strip()


def get_current_version() -> str:
    with open(MAIN_FILE) as fd:
        for line in fd:
            if m := VERSION_LINE.match(line.strip()):
                version = m.group(1)
                return version if version.startswith('v') else 'v' + version

    raise Exception('Unable to determine version')


def tag(version: str):
    tags = sh('git tag')
    tags_list = tags.split('\n') if tags else []
    versions = set(tags_list)
    if version in versions:
        print(f'A tag for version {version} already exists')
        return

    sh(f'git tag {version}')
    sh('git push --tags')


def release_exists(version: str):
    cmd = shlex.split(f'gh release view {version}')
    proc = subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    return proc.returncode == 0


def build(version: str) -> str:
    asset_base = os.environ['GITHUB_REPOSITORY'].split('/')[-1]
    platform = sh('uname')
    architecture = sh('uname -m')
    asset_name = f'{asset_base}-{version}-{platform}-{architecture}'.lower()
    home = os.environ['HOME']
    sh(f'go build -o {asset_name}',
       env={
           'CGO_ENABLED': '0',
           'GOPATH': os.path.join(home, 'go'),
           'GOCACHE': os.path.join(home, 'go', 'pkg', 'mod')
       })
    return asset_name


def release(version: str):
    if release_exists(version):
        print(f'A release already exists for version {version}')
        return

    asset_name = build(version)
    sh(f'gh release create -n "Release {version}" -t "{version}" "{version}"')
    sh(f'gh release upload "{version}" "{asset_name}"')


def tag_and_release():
    version = get_current_version()
    tag(version)
    release(version)


def main():
    tag_and_release()


if __name__ == '__main__':
    main()
