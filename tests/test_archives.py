import dataclasses
import os
import shlex
import shutil
import subprocess
import unittest

import yaml

ARTIFACTS_DIR = 'out'
BASE_CONFIG = {
    'archives': [],
    'settings': {
        'bin_dir': f'{ARTIFACTS_DIR}/bin',
        'extract_dir': f'{ARTIFACTS_DIR}/ext',
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


TESTS_CASES = [
    ArchiveTest(
        name='tar_archive_no_root_dir',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-no-root-dir.tar',
        ),
        symlink=Symlink(
            link_name='bin/foo',
            target='ext/foo/foo'
        ),
    ),
    ArchiveTest(
        name='tar_archive_root_dir_different_than_exec',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.tar',
        ),
        symlink=Symlink(
            link_name='bin/foo',
            target='ext/foo-1.2.3-amd64/foo'
        ),
    ),
    ArchiveTest(
        name='tar_archive_root_dir_same_as_exec',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.tar',
        ),
        symlink=Symlink(
            link_name='bin/foo',
            target='ext/foo/foo'
        ),
    ),
    ArchiveTest(
        name='zip_archive_no_root_dir',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-no-root-dir.zip',
        ),
        symlink=Symlink(
            link_name='bin/foo',
            target='ext/foo/foo'
        ),
    ),
    ArchiveTest(
        name='zip_archive_root_dir_different_than_exec',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.zip',
        ),
        symlink=Symlink(
            link_name='bin/foo',
            target='ext/foo-1.2.3-amd64/foo'
        ),
    ),
    ArchiveTest(
        name='zip_archive_root_dir_same_as_exec',
        archive=Archive(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.zip',
        ),
        symlink=Symlink(
            link_name='bin/foo',
            target='ext/foo/foo'
        ),
    )
]


class BaseTestCase(unittest.TestCase):
    def tearDown(self):
        if os.path.exists(ARTIFACTS_DIR):
            shutil.rmtree(ARTIFACTS_DIR)


def gen_test(test_case: ArchiveTest):
    def test(self):
        archives = [test_case.archive.__dict__]
        config = BASE_CONFIG | {'archives': archives}

        if not os.path.exists(ARTIFACTS_DIR):
            os.mkdir(ARTIFACTS_DIR)
        with open(f'{ARTIFACTS_DIR}/fup.yml', 'wt') as fd:
            yaml.dump(config, Dumper=yaml.SafeDumper, stream=fd)

        cmd = shlex.split(f'fup -p archive -f {ARTIFACTS_DIR}/fup.yml -b')
        proc = subprocess.run(cmd)
        self.assertTrue(proc.returncode == 0)

        link_name = os.path.join(os.path.realpath(ARTIFACTS_DIR), test_case.symlink.link_name)
        target = os.path.join(os.path.realpath(ARTIFACTS_DIR), test_case.symlink.target)
        self.assertEqual(os.path.realpath(link_name), target)

    return test


if __name__ == '__main__':
    for case in TESTS_CASES:
        test_method = gen_test(case)
        setattr(BaseTestCase, f'test_{case.name}', test_method)
    unittest.TextTestRunner().run(unittest.TestLoader().loadTestsFromTestCase(BaseTestCase))
