import os
import requests
import tarfile
import uuid

from pyinfra.api import FunctionCommand, operation
from pyinfra.operations import files

import tasks.config
from typing import Dict


def expand(url: str, context: Dict[str, str]):
    cur_dlr_index = -1
    parsing_var = False

    varmap = {}
    cur_var = ''

    for i, c in enumerate(url):
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
        url = url.replace(f'${{{var}}}', str(val))

    return url


import http.client
import re
import urllib

CHUNK_SIZE = 8192
CONTENT_DISPOSITION_FILENAME_REGEX = re.compile(r'filename=(.*)')


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
    url = urllib.parse.urlparse(url)
    conn = get_connection(url)
    path = f'{url.path}'
    if url.query:
        path += '?{url.query}'
    if url.fragment:
        path += '#{url.fragment}'
    conn.request(method, path)
    resp = conn.getresponse()

    with open(output_file, 'wb') as o:
        while chunk := resp.read(CHUNK_SIZE):
            o.write(chunk)


def download(url, target):
    http_request(url, 'GET', target)


def do_extract_archive(state, host, url, dest):
    tmpfile = f'/tmp/{uuid.uuid4()}'
    download(url, tmpfile)
    with tarfile.open(tmpfile) as tf:
        tf.extractall(dest)
    os.unlink(tmpfile)


@operation
def extract_archive(url=None, extract_dir=None, state=None, host=None):
    yield FunctionCommand(do_extract_archive, [url, extract_dir], {})


def extract(cfg):
    urls = []
    extract_dir = os.path.expanduser(cfg.settings.archive_dir)
    for archive in cfg.archives:
        url = archive['url']
        url = expand(url, cfg.settings.vars)
        urls.append(url)
        extract_archive(url=url, extract_dir=extract_dir)
