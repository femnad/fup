import os

from pyinfra.operations import files

from tasks.config import Template

FILE_PERMISSIONS = '0755'
TEMPLATED_FILES_SUFFIX = 'files'


def template_file(template: Template, should_sudo):
    src = f'{TEMPLATED_FILES_SUFFIX}/{template.src}'
    dest = os.path.expanduser(template.dest)

    context = {k: os.getenv(v) for k, v in template.context.items()}

    files.template(src=src, dest=dest, mode=FILE_PERMISSIONS, _sudo=should_sudo, **context)


def run(config):
    home = os.getenv('HOME')

    for template in config.templates:
        template = Template(**template)
        should_sudo = not template.dest.startswith(home)
        template_file(template, should_sudo)
