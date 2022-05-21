import mimetypes
import os
import tarfile
import uuid
import zipfile

from pyinfra.api import FunctionCommand, operation

import tasks.config
import tasks.context
import tasks.http
import tasks.unless

UNLESS_TYPES = [tasks.unless.UnlessCmd, tasks.unless.UnlessFile]


def extract_tar(f, dest):
    with tarfile.open(f) as tf:
        tf.extractall(dest)


def extract_zip(f, dest):
    with zipfile.ZipFile(f) as zf:
        zf.extractall(dest)


EXTRACTORS = {
    'application/x-tar': extract_tar,
    'application/zip': extract_zip,
}


def get_extractor(url):
    url = tasks.http.resolve_redirect(url, 'GET')
    file_type = mimetypes.guess_type(url)

    if file_type is None or file_type[0] is None:
        raise Exception(f'Cannot determine file type for {url}')

    file_type = file_type[0]
    if file_type not in EXTRACTORS:
        raise Exception(f'Unable to extract {file_type}')

    return EXTRACTORS[file_type]


def do_extract_archive(archive: tasks.config.Archive, dest):
    url = archive.url
    tmpfile = f'/tmp/{uuid.uuid4()}'
    tasks.http.download(url, tmpfile)

    extractor_fn = get_extractor(url)
    extractor_fn(tmpfile, dest)

    if archive:
        pass

    os.unlink(tmpfile)


def extract_archive(archive, extract_dir):
    do_extract_archive(archive, extract_dir)


def do_get_unless(unless, cls):
    try:
        return cls(**unless)
    except TypeError:
        return


def should_extract_cmd(archive: tasks.config.Archive, _, unless: tasks.unless.UnlessCmd):
    return unless.should_proceed(archive.version)


def should_extract_ls(archive: tasks.config.Archive, settings, unless: tasks.unless.UnlessFile):
    context = {k: v for k, v in archive.__dict__.items() if isinstance(v, str)}
    context.update(settings.__dict__)
    return unless.should_proceed(context)


UNLESS_FN_MAPPING = {
    tasks.unless.UnlessCmd: should_extract_cmd,
    tasks.unless.UnlessFile: should_extract_ls,
}


def get_unless(unless):
    if not unless:
        return

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
    archive.url = tasks.context.expand(archive.url, var_map)
    archive.symlink = [tasks.context.expand(s, var_map) for s in archive.symlink]
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


def change_file_mode(filename, mode):
    os.chmod(filename, 0o755)


@operation
def extract(archive: tasks.config.Archive, settings: tasks.config.Settings, extract_dir: str):
    bin_file = None
    archive = tasks.config.Archive(**archive)
    archive = expand_archive(archive)

    if not should_extract(archive, settings):
        return

    if archive.binary:
        extract_dir = os.path.expanduser(f'{extract_dir}/{archive.binary}')
        bin_file = f'{extract_dir}/{archive.binary}'
        archive.symlink = [f'{archive.binary}/{archive.binary}']

    yield FunctionCommand(extract_archive, [archive, extract_dir], {})
    if archive.symlink:
        yield FunctionCommand(symlink_archive, [archive, settings.archive_dir], {})

    if bin_file:
        yield FunctionCommand(change_file_mode, [bin_file, 0o755], {})


def run(config):
    extract_dir = os.path.expanduser(config.settings.archive_dir)
    for archive in config.archives:
        extract(archive, config.settings, extract_dir)
