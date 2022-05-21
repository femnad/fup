import http.client
import os
import re
import urllib

CHUNK_SIZE = 8192
CONTENT_DISPOSITION_FILENAME_REGEX = re.compile(r'filename=(.*)')
USER_AGENT = 'github.com/femnad/fup'


def get_filename_from_content_disposition(content_disposition):
    if not content_disposition:
        return

    if match := CONTENT_DISPOSITION_FILENAME_REGEX.match(content_disposition):
        return match.group(1)


def get_connection(parsed_url: urllib.parse.ParseResult):
    if parsed_url.scheme == 'https':
        return http.client.HTTPSConnection(parsed_url.netloc)
    return http.client.HTTPConnection(parsed_url.netloc)


def http_request(url, method, output_file=None):
    parsed_url = urllib.parse.urlparse(url)
    conn = get_connection(parsed_url)

    path = f'{parsed_url.path}'
    if parsed_url.query:
        path += f'?{parsed_url.query}'
    if parsed_url.fragment:
        path += f'#{parsed_url.fragment}'

    conn.request(method, path, headers={'User-Agent': USER_AGENT})
    resp = conn.getresponse()

    if resp.status == 302:
        redirect_url = resp.headers['Location']
        http_request(redirect_url, method, output_file)
        return
    elif resp.status != 200:
        body = resp.read().decode('utf-8')
        raise Exception(f'Error during HTTP request to {url}: {resp.status} {body}')

    if output_file:
        output_dir = os.path.dirname(output_file)
        if not os.path.exists(output_dir):
            os.makedirs(output_dir)
    else:
        buffer = ''
        while chunk := resp.read(CHUNK_SIZE):
            buffer += chunk.decode('utf-8')
        return buffer

    with open(output_file, 'wb') as o:
        while chunk := resp.read(CHUNK_SIZE):
            o.write(chunk)


def download(url, target):
    http_request(url, 'GET', target)
