import os
import socket

from pyinfra.api import FactBase


class GitReady(FactBase):

    hostname = socket.gethostname()
    private_key = os.path.expanduser(f'~/.ssh/{hostname}')
    command = f'find {private_key} || true'

    def process(self, output):
        return len(output) > 0


class SshPullReady(FactBase):

    command = 'pass'

    def process(self, output):
        return len(output) > 0


class NeovimReady(FactBase):
    command = 'find ~/.config/nvim/plugged'

    def process(self, output):
        return len(output) > 0


class InTmux(FactBase):
    command = 'echo $TMUX'

    def process(self, output):
        return output[0].startswith('/tmp/tmux')


class IsLaptop(FactBase):
    command = 'acpi'

    def process(self, output):
        return not output[0].endswith('rate information unavailable')
