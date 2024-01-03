import dataclasses
import os
import shutil
import sys
import unittest

import fup_test

ABSOLUTE_ARTIFACTS_DIR = os.path.expanduser('~/out')
RELATIVE_ARTIFACTS_DIR = 'out'
ABSOLUTE_CONFIG = {
    'settings': {
        'bin_dir': f'{ABSOLUTE_ARTIFACTS_DIR}/bin',
        'release_dir': f'{ABSOLUTE_ARTIFACTS_DIR}/ext',
    }
}
RELATIVE_CONFIG = {
    'settings': {
        'bin_dir': f'{RELATIVE_ARTIFACTS_DIR}/bin',
        'release_dir': f'{RELATIVE_ARTIFACTS_DIR}/ext',
    }
}


@dataclasses.dataclass
class Release:
    name: str
    url: str
    target: str = ''


@dataclasses.dataclass
class Symlink:
    link_name: str
    target: str


@dataclasses.dataclass
class ReleaseTest:
    release: Release
    name: str
    symlink: Symlink
    config: dict = dataclasses.field(default_factory=lambda: RELATIVE_CONFIG)
    relative: bool = True


TESTS_CASES = [
    ReleaseTest(
        name='tar_archive_no_root_dir',
        release=Release(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-no-root-dir.tar',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo/foo'),
    ),
    ReleaseTest(
        name='tar_archive_root_dir_different_than_exec',
        release=Release(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.tar',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo-1.2.3-amd64/foo'),
    ),
    ReleaseTest(
        name='tar_archive_root_dir_different_than_exec_override_target',
        release=Release(
            name='foo',
            target='fred',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.tar',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/fred/foo'),
    ),
    ReleaseTest(
        name='tar_archive_root_dir_same_as_exec',
        release=Release(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.tar',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo/foo'),
    ),
    ReleaseTest(
        name='tar_archive_root_dir_same_as_exec_override_target',
        release=Release(
            name='foo',
            target='qux',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.tar',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/qux/foo'),
    ),
    ReleaseTest(
        name='zip_archive_no_root_dir',
        release=Release(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-no-root-dir.zip',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo/foo'),
    ),
    ReleaseTest(
        name='zip_archive_root_dir_different_than_exec',
        release=Release(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.zip',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo-1.2.3-amd64/foo'),
    ),
    ReleaseTest(
        name='zip_archive_root_dir_different_than_exec_override_target',
        release=Release(
            name='foo',
            target='baz',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-different-than-exec.zip',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/baz/foo'),
    ),
    ReleaseTest(
        name='zip_archive_root_dir_same_as_exec',
        release=Release(
            name='foo',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.zip',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/foo/foo'),
    ),
    ReleaseTest(
        name='zip_archive_root_dir_same_as_exec_override_target',
        release=Release(
            name='foo',
            target='bar',
            url='https://github.com/femnad/fup/releases/download/test-payload/release-root-dir-same-as-exec.zip',
        ),
        symlink=Symlink(link_name='bin/foo', target='ext/bar/foo'),
    ),
    ReleaseTest(
        name='zip_archive_root_dir_same_as_exec_abs_dirs',
        release=Release(
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


def gen_test(test_case: ReleaseTest):

    def test(self):
        releases = [test_case.release.__dict__]
        config = test_case.config | {'release': releases}

        artifacts_dir = RELATIVE_ARTIFACTS_DIR if test_case.relative else ABSOLUTE_ARTIFACTS_DIR

        return_code = fup_test.run_fup(config, f'{artifacts_dir}/fup.yml', 'release')
        self.assertTrue(return_code == 0)

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
