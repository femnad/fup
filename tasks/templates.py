import os

from pyinfra.api import operation, FunctionCommand
from pyinfra.operations import files

from tasks.config import Template

TEMPLATED_FILES_SUFFIX = 'files'


def do_template(template, should_sudo):
    context = {k: os.getenv(v) for k, v in template.context.items()}
    files.template(src=template.src, dest=template.dest, mode=template.mode, _sudo=should_sudo, **context)


@operation
def template_file(template: Template, should_sudo):
    template.src = f'{TEMPLATED_FILES_SUFFIX}/{template.src}'
    yield FunctionCommand(do_template, [template, should_sudo], {})


def run(config):
    home = os.getenv('HOME')

    for template in config.templates:
        template = Template(**template)
        template.dest = os.path.expanduser(template.dest)
        should_sudo = not template.dest.startswith(home)

        template_file(template, should_sudo)
