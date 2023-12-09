import dataclasses
import os
import shlex
import shutil
import subprocess
import sys
import unittest

import yaml

ABSOLUTE_ARTIFACTS_DIR = os.path.expanduser('~/out')
RELATIVE_ARTIFACTS_DIR = 'out'
ABSOLUTE_CONFIG = {
    'settings': {
        'bin_dir': f'{ABSOLUTE_ARTIFACTS_DIR}/bin',
        'extract_dir': f'{ABSOLUTE_ARTIFACTS_DIR}/ext',
    }
}
RELATIVE_CONFIG = {
    'settings': {
        'bin_dir': f'{RELATIVE_ARTIFACTS_DIR}/bin',
        'extract_dir': f'{RELATIVE_ARTIFACTS_DIR}/ext',
    }
}


@dataclasses.dataclass
class Archive:
    name: str
    url: str


@dataclasses.dataclass
class Symlink:
    link_name: str
    target: str


@dataclasses.dataclass
class ArchiveTest:
    archive: Archive
    name: str
    symlink: Symlink
    config: dict = dataclasses.field(default_factory=lambda: RELATIVE_CONFIG)
    relative: bool = True


TESTS_CASES = [
    ArchiveTest(
        name='tar_archive_no_root_dir',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-no-root-dir.tar',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo/foo'),
    ),
    ArchiveTest(
        name='tar_archive_root_dir_different_than_exec',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.tar',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo-1.2.3-amd64/foo'),
    ),
    ArchiveTest(
        name='tar_archive_root_dir_same_as_exec',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.tar',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo/foo'),
    ),
    ArchiveTest(
        name='zip_archive_no_root_dir',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-no-root-dir.zip',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo/foo'),
    ),
    ArchiveTest(
        name='zip_archive_root_dir_different_than_exec',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.zip',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo-1.2.3-amd64/foo'),
    ),
    ArchiveTest(
        name='zip_archive_root_dir_same_as_exec',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.zip',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo/foo'),
    ),
    ArchiveTest(
        name='zip_archive_root_dir_same_as_exec_abs_dirs',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.zip',
        ),
        symlink=Symlink(link_name=f'{ABSOLUTE_ARTIFACTS_DIR}/bin/foo', target=f'{ABSOLUTE_ARTIFACTS_DIR}/ext/foo/foo'),
        config=ABSOLUTE_CONFIG,
        relative=False,
    ),
]


class BaseTestCase(unittest.TestCase):

    def tearDown(self):
        for path in [ABSOLUTE_ARTIFACTS_DIR, RELATIVE_ARTIFACTS_DIR]:
            if os.path.exists(path):
                shutil.rmtree(path)


def ensure_abs(path: str, root: str) -> str:
    if os.path.isabs(path):
        return path

    return os.path.join(os.path.realpath(root), path)


def gen_test(test_case: ArchiveTest):

    def test(self):
        archives = [test_case.archive.__dict__]
        config = test_case.config | {'archives': archives}

        artifacts_dir = RELATIVE_ARTIFACTS_DIR if test_case.relative else ABSOLUTE_ARTIFACTS_DIR

        if not os.path.exists(artifacts_dir):
            os.mkdir(artifacts_dir)
        with open(f'{artifacts_dir}/fup.yml', 'wt') as fd:
            yaml.dump(config, Dumper=yaml.SafeDumper, stream=fd)

        go_path = os.getenv('GOPATH', os.path.expanduser('~/go'))
        fup_bin = os.path.join(go_path, 'bin', 'fup')
        cmd = shlex.split(f'{fup_bin} -p archive -f {artifacts_dir}/fup.yml -l 0 -n')
        proc = subprocess.run(cmd)
        self.assertTrue(proc.returncode == 0)

        link_name = ensure_abs(test_case.symlink.link_name, artifacts_dir)
        target = ensure_abs(test_case.symlink.target, artifacts_dir)
        self.assertEqual(os.path.realpath(link_name), target)

    return test


if __name__ == '__main__':
    for case in TESTS_CASES:
        test_method = gen_test(case)
        setattr(BaseTestCase, f'test_{case.name}', test_method)

    if not unittest.TextTestRunner().run(unittest.TestLoader().loadTestsFromTestCase(BaseTestCase)).wasSuccessful():
        sys.exit(1)
