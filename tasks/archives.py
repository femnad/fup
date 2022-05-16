import http.client
import os
import re
import subprocess
import tarfile
import uuid
import urllib

from pyinfra.api import FunctionCommand, operation
from pyinfra.operations import files

import tasks.config
from typing import Dict

CHUNK_SIZE = 8192
CONTENT_DISPOSITION_FILENAME_REGEX = re.compile(r'filename=(.*)')
UNLESS_TYPES = [tasks.config.UnlessCmd, tasks.config.UnlessFile]


def expand(s: str, context: Dict[str, str]):
    cur_dlr_index = -1
    parsing_var = False

    varmap = {}
    cur_var = ''

    for i, c in enumerate(s):
        if c == '$':
            cur_dlr_index = i
            continue
        elif c == '{' and i == cur_dlr_index + 1:
            parsing_var = True
            continue
        elif c == '}':
            parsing_var = False
            value = context[cur_var]
            varmap[cur_var] = value
            cur_var = ''
            cur_dlr_index = -1
        elif parsing_var:
            cur_var += c

    for var, val in varmap.items():
        s = s.replace(f'${{{var}}}', str(val))

    return s


def get_filename_from_content_disposition(content_disposition):
    if not content_disposition:
        return

    if match := CONTENT_DISPOSITION_FILENAME_REGEX.match(content_disposition):
        return match.group(1)


def get_connection(parsed_url: urllib.parse.ParseResult):
    if parsed_url.scheme == 'https':
        return http.client.HTTPSConnection(parsed_url.netloc)
    return http.client.HTTPConnection(parsed_url.netloc)


def http_request(url, method, output_file):
    parsed_url = urllib.parse.urlparse(url)
    conn = get_connection(parsed_url)

    path = f'{parsed_url.path}'
    if parsed_url.query:
        path += f'?{parsed_url.query}'
    if parsed_url.fragment:
        path += f'#{parsed_url.fragment}'

    conn.request(method, path)
    resp = conn.getresponse()

    if resp.status == 302:
        redirect_url = resp.headers['Location']
        http_request(redirect_url, method, output_file)
        return
    elif resp.status != 200:
        body = resp.read().decode('utf-8')
        raise Exception(f'Error during HTTP request to {url}: {resp.status} {body}')

    with open(output_file, 'wb') as o:
        while chunk := resp.read(CHUNK_SIZE):
            o.write(chunk)


def download(url, target):
    http_request(url, 'GET', target)


def do_extract_archive(url, dest):
    tmpfile = f'/tmp/{uuid.uuid4()}'
    download(url, tmpfile)
    with tarfile.open(tmpfile) as tf:
        tf.extractall(dest)
    os.unlink(tmpfile)


@operation
def extract_archive(url=None, extract_dir=None):
    yield FunctionCommand(do_extract_archive, [url, extract_dir], {})


def get_fn(operation: str, parameter: int):
    if operation == 'head':
        return lambda x: x.split('\n')[parameter]
    elif operation == 'split':
        return lambda x: x.split()[parameter]
    else:
        raise Exception(f'Unknown operation {operation}')


def get_version(output, version_fn):
    ops = []

    for op in version_fn.split('|'):
        operation, parameter = op.strip().split()
        parameter = int(parameter)
        ops.append(get_fn(operation, parameter))

    for op in ops:
        output = op(output)

    return output


def do_get_unless(unless, cls):
    try:
        return cls(**unless)
    except TypeError:
        return


def should_extract_cmd(archive: tasks.config.Archive, _, unless: tasks.config.UnlessCmd):
    proc = subprocess.run(unless.cmd, shell=True, capture_output=True, text=True)
    if proc.returncode != 0:
        return True

    if not unless.post:
        return False

    output = proc.stdout.strip()
    current_version = get_version(output, unless.post)
    if current_version == archive.version:
        return False

    return True


def should_extract_ls(archive: tasks.config.Archive, settings, unless: tasks.config.UnlessFile):
    context = {k: v for k, v in archive.__dict__.items() if isinstance(v, str)}
    context.update(settings.__dict__)
    ls_target = expand(unless.ls, context)
    ls_target = os.path.expanduser(ls_target)
    return not os.path.exists(ls_target)


UNLESS_FN_MAPPING = {
    tasks.config.UnlessCmd: should_extract_cmd,
    tasks.config.UnlessFile: should_extract_ls,
}


def get_unless(unless):
    for unless_type in UNLESS_TYPES:
        if found_unless := do_get_unless(unless, unless_type):
            return found_unless

    raise Exception(f'Cannot determine unless type for {unless}')


def should_extract(archive: tasks.config.Archive, settings):
    if not archive.unless:
        return True

    unless = get_unless(archive.unless)
    unless_fn = UNLESS_FN_MAPPING[type(unless)]

    return unless_fn(archive, settings, unless)


def expand_archive(archive: tasks.config.Archive):
    var_map = {'version': archive.version}
    archive.url = expand(archive.url, var_map)
    archive.symlink = [expand(s, var_map) for s in archive.symlink]
    return archive


def symlink_archive(archive: tasks.config.Archive, archive_dir: str):
    if not archive.symlink:
        return

    for src in archive.symlink:
        name = os.path.basename(src)
        src = os.path.join(archive_dir, src)
        src = os.path.expanduser(src)
        dst = os.path.expanduser(f'~/bin/{name}')

        # Remove if target is a broken symlink
        if os.path.lexists(dst):
            os.unlink(dst)

        os.symlink(src, dst)


def extract(cfg):
    extract_dir = os.path.expanduser(cfg.settings.archive_dir)
    for archive in cfg.archives:
        archive = expand_archive(archive)

        if not should_extract(archive, cfg.settings):
            continue

        extract_archive(url=archive.url, extract_dir=extract_dir)
        symlink_archive(archive, cfg.settings.archive_dir)
