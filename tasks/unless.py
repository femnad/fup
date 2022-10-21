from dataclasses import dataclass
import os
import re
import subprocess
from typing import Dict

import tasks.context


@dataclass
class UnlessFile:
    ls: str

    def should_proceed(self, context: Dict = None):
        if not context:
            context = {}

        ls_target = tasks.context.expand(self.ls, context)
        ls_target = os.path.expanduser(ls_target)
        return not os.path.exists(ls_target)


def contains(output, needle):
    for line in output.split('\n'):
        if needle in line:
            return True

    return False


def split(s, index):
    split = s.split()
    index = int(index)
    if len(split) < index + 1:
        raise Exception(f"Invalid split for string {s} and index {index}")
    return split[index]


@dataclass
class UnlessCmd:
    cmd: str
    post: str = None

    POST_FNS = {
        'cut': lambda x, p: x[int(p):],
        'head': lambda x, p: x.split('\n')[int(p)],
        'split': split,
        'contains': contains,
        'match': lambda x, p: re.match(x, p),
    }

    def get_fn(self, operation: str, parameter: int):
        if operation not in self.POST_FNS:
            raise Exception(f'Unknown operation {operation}')

        return lambda x: self.POST_FNS[operation](x, parameter)

    def get_version(self, output, version_fn):
        ops = []

        for op in version_fn.split('|'):
            operation, parameter = op.strip().split()
            ops.append(self.get_fn(operation, parameter))

        for op in ops:
            output = op(output)

        return output

    def should_proceed(self, version: str = ''):
        proc = subprocess.run(self.cmd, shell=True, capture_output=True, text=True)
        if proc.returncode != 0:
            return True

        if not self.post:
            return False

        if not version:
            return

        output = proc.stdout.strip()
        current_version = self.get_version(output, self.post)
        if current_version == version:
            return False

        return True
