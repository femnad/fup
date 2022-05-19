from typing import List, Union

from pyinfra.api import StringCommand, operation

from tasks.config import Service
from tasks.recipes import run_command


def get_systemctl_command(service, command, readonly=False):
    maybe_user = ' --user' if service.user else ''
    maybe_sudo = '' if service.user or readonly else 'sudo '

    return f'{maybe_sudo}systemctl{maybe_user} {command} {service.name}'


def maybe_actuate(service: Service, check_cmd: str, actuation_cmd: str, actuated_state: str) -> Union[None, str]:
    if not getattr(service, actuation_cmd):
        return

    is_actuated_cmd = get_systemctl_command(service, check_cmd, readonly=True)
    if run_command(is_actuated_cmd, raise_on_error=False) == actuated_state:
        return

    return get_systemctl_command(service, actuation_cmd)


def maybe_enable(service: Service):
    return maybe_actuate(service, 'is-enabled', 'enable', 'enabled')


def maybe_start(service: Service):
    return maybe_actuate(service, 'is-active', 'start', 'active')


def init_service(service: Service):
    if enable_cmd := maybe_enable(service):
        yield StringCommand(enable_cmd)

    if start_cmd := maybe_start(service):
        yield StringCommand(start_cmd)


@operation
def init_services(services: List[Service]):
    for service in services:
        yield from init_service(service)


def run(config):
    services = [Service(**s) for s in config.services]
    init_services(services)
