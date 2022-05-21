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

    command = 'find ~/.password-store || true'

    def process(self, output):
        return len(output) > 0


class NeovimReady(FactBase):
    command = 'find ~/.config/nvim/plugged || true'

    def process(self, output):
        return len(output) > 0


class InTmux(FactBase):
    command = 'echo $TMUX'

    def process(self, output):
        return output[0].startswith('/tmp/tmux')


class IsLaptop(FactBase):
    command = 'acpi || true'

    def process(self, output):
        if not output:
            return False

        return not output[0].endswith('rate information unavailable')


class IsFedora(FactBase):
    command = 'cat /etc/os-release'

    def process(self, output):
        return output[2].split('=')[-1] == 'fedora'
