import os
import socket

from pyinfra.api import FactBase


class GitReady(FactBase):

    hostname = socket.gethostname()
    private_key = os.path.expanduser(f'~/.ssh/{hostname}')
    command = f'find {private_key} || true'

    def process(self, output):
        return len(output) > 0


class SshReady(FactBase):

    hostname = socket.gethostname()
    command = f'ssh-add -l | grep {hostname} || true'

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


class IsOS(FactBase):
    command = 'cat /etc/os-release'

    def process(self, output):
        for l in output:
            if l == f'ID={self.os}':
                return True

        return False


class IsFedora(IsOS):
    os = 'fedora'


class IsUbuntu(IsOS):
    os = 'ubuntu'
