import dataclasses
import os
import shutil
import sys
import unittest

import fup_test

TEST_DIR = 'lineinfile'


@dataclasses.dataclass
class Replace:
    old: str
    new: str = ''
    absent: bool = False
    ensure: bool = False
    regex: bool = False


@dataclasses.dataclass
class LineInFile:
    name: str
    file: str
    replace: list[Replace]
    when: str = ''


@dataclasses.dataclass
class LineInFileTest:
    line_in_file: LineInFile
    file: dict[str, str]
    name: str
    want: dict[str, str]


TESTS_CASES = [
    LineInFileTest(name='Exact match',
                   file={
                       'baz': '''foo
bar
baz
''',
                   },
                   want={'baz': '''foo
qux
baz
'''},
                   line_in_file=LineInFile(name='replace',
                                           file=f'{TEST_DIR}/baz',
                                           replace=[Replace(old='bar', new='qux')])),
    LineInFileTest(name='No matches',
                   file={
                       'baz': '''foo
bar
baz
''',
                   },
                   want={'baz': '''foo
bar
baz
'''},
                   line_in_file=LineInFile(name='replace',
                                           file=f'{TEST_DIR}/baz',
                                           replace=[Replace(old='qux', new='fred')])),
    LineInFileTest(name='Regex match',
                   file={
                       'baz': '''foo
fred984
baz
''',
                   },
                   want={'baz': '''foo
barney
baz
'''},
                   line_in_file=LineInFile(name='replace',
                                           file=f'{TEST_DIR}/baz',
                                           replace=[Replace(old='fred[0-9]+', new='barney', regex=True)])),
    LineInFileTest(name='Absent',
                   file={
                       'baz': '''foo
bar
baz
''',
                   },
                   want={'baz': '''foo
baz
'''},
                   line_in_file=LineInFile(name='replace',
                                           file=f'{TEST_DIR}/baz',
                                           replace=[Replace(old='bar', absent=True)])),
    LineInFileTest(name='Ensure',
                   file={
                       'baz': '''foo
baz
''',
                   },
                   want={'baz': '''foo
baz
bar
'''},
                   line_in_file=LineInFile(name='replace',
                                           file=f'{TEST_DIR}/baz',
                                           replace=[Replace(old='stuff', new='bar', ensure=True)])),
]


class BaseTestCase(unittest.TestCase):

    def tearDown(self):
        if os.path.exists(TEST_DIR):
            shutil.rmtree(TEST_DIR)


def ensure_abs(path: str, root: str) -> str:
    if os.path.isabs(path):
        return path

    return os.path.join(os.path.realpath(root), path)


def gen_test(test_case: LineInFileTest):

    def test(self):
        lines = [dataclasses.asdict(test_case.line_in_file)]
        config = {'line': lines}

        for file, want_content in test_case.file.items():
            file = os.path.join(TEST_DIR, file)
            file_dir = os.path.dirname(file)
            if not os.path.exists(file_dir):
                os.makedirs(file_dir)

            with open(file, 'w') as fd:
                fd.write(want_content)

        return_code = fup_test.run_fup(config, f'{TEST_DIR}/fup.yml', 'line')
        self.assertTrue(return_code == 0)

        for file, want_content in test_case.want.items():
            file = os.path.join(TEST_DIR, file)
            with open(file) as fd:
                got = fd.read()
                self.assertEqual(got, want_content)

    return test


if __name__ == '__main__':
    for case in TESTS_CASES:
        test_method = gen_test(case)
        setattr(BaseTestCase, f'test_{case.name}', test_method)

    if not unittest.TextTestRunner().run(unittest.TestLoader().loadTestsFromTestCase(BaseTestCase)).wasSuccessful():
        sys.exit(1)
