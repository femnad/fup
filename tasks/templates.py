from dataclasses import dataclass
from hashlib import sha1
import os
import uuid
from typing import Union

from pyinfra.api import operation, FunctionCommand
from pyinfra.api.util import get_template

from tasks.config import Template
from tasks.recipes import run_command

HASH_READ_BUFFER = 8192
TEMPLATED_FILES_SUFFIX = 'files'


@dataclass
class UpdateOp:
    output: str
    template: Template
    should_update_file: bool
    should_update_mode: bool


def get_temp_file():
    return f'/tmp/{uuid.uuid4()}'


def get_file_hash(filename):
    h = sha1()

    with open(filename, 'rb') as f:
        while b := f.read(HASH_READ_BUFFER):
            h.update(b)

    return h.hexdigest()


def get_hash(s: str):
    h = sha1()
    h.update(s.encode('utf-8'))
    return h.hexdigest()


def get_mode(filename):
    mode = os.stat(filename).st_mode
    return str(oct(mode))[-4:]


def update_mode(filename, mode, should_sudo):
    run_command(f'chmod {mode} {filename}', sudo=should_sudo)


def update_file(temp_file, output, dest, should_sudo):
    with open(temp_file, 'w') as f:
        f.write(output)

    run_command(f'cp {temp_file} {dest}', sudo=should_sudo)


def remove_file(filename, should_sudo):
    run_command(f'rm {filename}', sudo=should_sudo)


def do_template_file(update_op: UpdateOp):
    template = update_op.template
    dest = template.dest
    output = update_op.output

    home = os.getenv('HOME')
    should_sudo = not template.dest.startswith(home)

    if update_op.should_update_file:
        temp_file = get_temp_file()
        yield FunctionCommand(update_file, [temp_file, output, dest, should_sudo], {})
        yield FunctionCommand(remove_file, [temp_file, should_sudo], {})

    if update_op.should_update_mode:
        yield FunctionCommand(update_mode, [dest, template.mode, should_sudo], {})


def render_template(template: Template) -> str:
    context = {
        k: os.getenv(v[1:]) if isinstance(v, str) and v.startswith('$') else v
        for k, v in template.context.items()
    }
    return get_template(template.src).render(context)


def maybe_template_file(template: Template) -> Union[None, UpdateOp]:
    output = render_template(template)

    prev_hash = get_file_hash(template.dest)
    new_hash = get_hash(output)
    prev_mode = get_mode(template.dest)

    should_update_file = prev_hash != new_hash
    should_update_mode = prev_mode != template.mode

    if not (should_update_file or should_update_mode):
        return

    return UpdateOp(
        output=output,
        template=template,
        should_update_file=should_update_file,
        should_update_mode=should_update_mode,
    )


@operation
def template_file(template: Template):
    template.src = f'{TEMPLATED_FILES_SUFFIX}/{template.src}'

    if update_op := maybe_template_file(template):
        yield FunctionCommand(do_template_file, [update_op], {})


def run(config):
    for template in config.templates:
        template = Template(**template)
        template.dest = os.path.expanduser(template.dest)
        template_file(template)
