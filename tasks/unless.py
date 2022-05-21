from dataclasses import dataclass
import os
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


@dataclass
class UnlessCmd:
    cmd: str
    post: str = None

    def get_fn(self, operation: str, parameter: int):
        if operation == 'head':
            return lambda x: x.split('\n')[parameter]
        elif operation == 'split':
            return lambda x: x.split()[parameter]
        else:
            raise Exception(f'Unknown operation {operation}')

    def get_version(self, output, version_fn):
        ops = []

        for op in version_fn.split('|'):
            operation, parameter = op.strip().split()
            parameter = int(parameter)
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
