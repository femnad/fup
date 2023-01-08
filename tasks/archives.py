import mimetypes
import os
import tarfile
import uuid
import zipfile

from pyinfra.api import FunctionCommand, operation

import tasks.config
import tasks.context
import tasks.http
import tasks.ops
import tasks.unless
import tasks.when

UNLESS_TYPES = [tasks.unless.UnlessCmd, tasks.unless.UnlessFile]


def do_set_permissions(file_iterator):
    for f in file_iterator:
        if f.isdir():
            f.mode = 0o755
        elif f.isfile():
            f.mode = 0o644


def is_within_directory(directory: str, target: str) -> bool:
    abs_directory = os.path.abspath(directory)
    abs_target = os.path.abspath(target)
    prefix = os.path.commonprefix([abs_directory, abs_target])

    return prefix == abs_directory


def safe_extract(tar: tarfile.TarFile, path: str):
    for member in tar.getmembers():
        member_path = os.path.join(path, member.name)
        if not is_within_directory(path, member_path):
            raise Exception('Attempted path traversal in tar file')

    tar.extractall(path)


def extract_tar(file_name: str, dest: str, set_permissions=False) -> None:
    with tarfile.open(file_name) as tf:
        if set_permissions:
            do_set_permissions(tf)

        safe_extract(tf, dest)


def extract_zip(f, dest, set_permissions=False):
    with zipfile.ZipFile(f) as zf:
        if set_permissions:
            do_set_permissions(zf)
        zf.extractall(dest)


EXTRACTORS = {
    'application/x-tar': extract_tar,
    'application/zip': extract_zip,
}


def get_extractor(url):
    url = tasks.http.resolve_redirect(url, 'GET')
    file_type = mimetypes.guess_type(url)

    if file_type is None or file_type[0] is None:
        file_type = ('application/x-tar', 'gzip')

    file_type = file_type[0]
    assert file_type is not None

    if file_type not in EXTRACTORS:
        raise Exception(f'Unable to extract {file_type}')

    return EXTRACTORS[file_type]


def do_extract_archive(archive: tasks.config.Archive, dest):
    url = archive.url
    if not url:
        raise Exception(f'No URL given for archive {archive}')

    tmpfile = f'/tmp/{uuid.uuid4()}'
    tasks.http.download(url, tmpfile)

    extractor_fn = get_extractor(url)
    extractor_fn(tmpfile, dest, archive.set_permissions)

    os.unlink(tmpfile)


def maybe_execute_after(archive, extract_dir):
    if not archive.execute_after:
        return

    cmds = archive.execute_after.strip().split('\n')
    cmds = [tasks.context.expand(c, {'version': archive.version}) for c in cmds]
    tasks.ops.run_commands(cmds, pwd=extract_dir)


def extract_archive(archive, extract_dir):
    do_extract_archive(archive, extract_dir)
    maybe_execute_after(archive, extract_dir)


def do_get_unless(unless, cls):
    try:
        return cls(**unless)
    except TypeError:
        return


def should_extract_cmd(archive: tasks.config.Archive, _, unless: tasks.unless.UnlessCmd):
    return unless.should_proceed(archive.version)


def should_extract_ls(archive: tasks.config.Archive, settings, unless: tasks.unless.UnlessFile):
    context = {k: v for k, v in archive.__dict__.items() if type(v) in [int, float, str]}
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

    if not tasks.when.should_run(archive.when):
        return

    if not should_extract(archive, settings):
        return

    if archive.binary:
        extract_dir = os.path.expanduser(f'{extract_dir}/{archive.binary}')
        bin_file = f'{extract_dir}/{archive.binary}'
        archive.symlink = [f'{archive.binary}/{archive.binary}']

    if archive.target:
        extract_dir = os.path.join(extract_dir, archive.target)

    yield FunctionCommand(extract_archive, [archive, extract_dir], {})
    if archive.symlink:
        yield FunctionCommand(symlink_archive, [archive, settings.archive_dir], {})

    if bin_file:
        yield FunctionCommand(change_file_mode, [bin_file, 0o755], {})


def run(config):
    extract_dir = os.path.expanduser(config.settings.archive_dir)
    for archive in config.archives:
        extract(archive, config.settings, extract_dir)
