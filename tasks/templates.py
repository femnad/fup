import os

from pyinfra.operations import files

from tasks.config import Template

TEMPLATED_FILES_SUFFIX = 'files'


def template_file(template: Template, should_sudo):
    src = f'{TEMPLATED_FILES_SUFFIX}/{template.src}'
    context = {k: os.getenv(v) for k, v in template.context.items()}

    files.template(src=src, dest=template.dest, mode=template.mode, _sudo=should_sudo, **context)


def run(config):
    home = os.getenv('HOME')

    for template in config.templates:
        template = Template(**template)
        template.dest = os.path.expanduser(template.dest)
        should_sudo = not template.dest.startswith(home)

        template_file(template, should_sudo)
