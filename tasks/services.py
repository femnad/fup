import copy
import os
from typing import List, Union

from pyinfra.api import FunctionCommand, StringCommand, operation

from tasks.config import Service, ServiceUnit, Template
from tasks.ops import run_command
from tasks.templates import do_template_file, maybe_template_file
import tasks.when

SERVICE_FILE_MODE = '0644'
SERVICE_TEMPLATE = 'files/service.j2'


def get_systemctl_command(service, command, readonly=False, no_service=False):
    maybe_user = ' --user' if service.user else ''
    maybe_sudo = '' if service.user or readonly else 'sudo '
    maybe_service = '' if no_service else f' {service.name}'

    return f'{maybe_sudo}systemctl{maybe_user} {command}{maybe_service}'


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


def get_service_file(service: Service):
    if service.user:
        return os.path.expanduser(f'~/.config/systemd/user/{service.name}.service')

    return f'/etc/systemd/system/{service.name}.service'


def maybe_daemon_reload(service, op):
    if op.changed:
        daemon_reload = get_systemctl_command(service, 'daemon-reload', no_service=True)
        yield StringCommand(daemon_reload)


def template_service(service: Service):
    dest = get_service_file(service)

    template = Template(src=SERVICE_TEMPLATE, dest=dest, mode=SERVICE_FILE_MODE, context=service.unit.__dict__)
    if update_op := maybe_template_file(template):
        do_template_file(update_op)
        daemon_reload = get_systemctl_command(service, 'daemon-reload', no_service=True)
        yield StringCommand(daemon_reload)


def init_service(service: Service):
    if service.unit:
        service.unit = ServiceUnit(**service.unit)
        yield from template_service(service)

    if enable_cmd := maybe_enable(service):
        yield StringCommand(enable_cmd)

    if start_cmd := maybe_start(service):
        yield StringCommand(start_cmd)


@operation
def init_services(services: List[Service]):
    for service in services:
        if not tasks.when.should_run(service.when):
            continue

        yield from init_service(service)


def run(config):
    services = [Service(**s) for s in config.services]
    init_services(services)
