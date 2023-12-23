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


def do_get_current_version() -> str:
    with open(MAIN_FILE) as fd:
        for line in fd:
            if m := VERSION_LINE.match(line.strip()):
                version = m.group(1)
                return version if version.startswith('v') else 'v' + version

    raise Exception('Unable to determine version')


def get_current_version() -> tuple[str, bool]:
    version = do_get_current_version()
    tags = sh('git tag')
    tags_list = tags.split('\n') if tags else []
    versions = set(tags_list)
    print(version)
    print(versions)
    return version, version in versions


def tag(new_tag: str):
    sh(f'git tag {new_tag}')
    sh('git push --tags')


def create_release(version: str):
    sh(f'gh release create -n "Release {version}" -t "{version}" "version"')


def release(version: str):
    repo_name = os.environ['GITHUB_REPOSITORY'].split('/')[-1]
    platform = sh('uname')
    architecture = sh('uname -m')
    asset_name = f'{repo_name}-{version}-{platform}-{architecture}'.lower()
    sh(f'go build -o {asset_name}', env={'CGO_ENABLED': '0'})
    sh(f'gh release upload "{version}" "{asset_name}"')


def tag_and_release():
    version, exists = get_current_version()
    if exists:
        print(f'A tag for version {version} already exists')
        return

    tag(version)
    create_release(version)
    release(version)


def main():
    tag_and_release()


if __name__ == '__main__':
    main()
