from dataclasses import dataclass
from hashlib import sha1
import os
import uuid

from pyinfra.api import operation, FunctionCommand
from pyinfra.api.util import get_template

from tasks.config import Template
from tasks.recipes import run_command

HASH_READ_BUFFER = 8192
TEMPLATED_FILES_SUFFIX = "files"


@dataclass
class OpResult:
    changed: bool = False


def get_temp_file():
    return f"/tmp/{uuid.uuid4()}"


def get_hash(filename):
    h = sha1()

    with open(filename, "rb") as f:
        while b := f.read(HASH_READ_BUFFER):
            h.update(b)

    return h.hexdigest()


def get_mode(filename):
    return os.stat(filename).st_mode


def update_mode(filename, mode, should_sudo):
    run_command(f"chmod {mode} {filename}", sudo=should_sudo)


def update_file(src, dest, should_sudo):
    run_command(f"cp {src} {dest}", sudo=should_sudo)


def remove_file(filename, should_sudo):
    run_command(f"rm {filename}", sudo=should_sudo)


def do_template_file(
    template: Template, temp_file: str, should_sudo: bool, op_result: OpResult
):
    context = {k: os.getenv(v) for k, v in template.context.items()}
    output = get_template(template.src).render(context)

    with open(temp_file, "w") as f:
        f.write(output)

    dest = template.dest

    prev_hash = get_hash(dest)
    new_hash = get_hash(temp_file)

    changed = False

    if prev_hash != new_hash:
        yield FunctionCommand(update_file, [temp_file, dest, should_sudo], {})
        changed = True

    prev_mode = get_mode(dest)
    new_mode = get_mode(temp_file)

    if prev_mode != new_mode:
        yield FunctionCommand(update_mode, [dest, template.mode, should_sudo], {})
        changed = True

    op_result.changed = changed

    yield FunctionCommand(remove_file, [temp_file, should_sudo], {})


def do_template(template):
    home = os.getenv("HOME")
    should_sudo = not template.dest.startswith(home)

    temp_file = get_temp_file()
    op_result = OpResult()

    do_template_file(template, temp_file, should_sudo, op_result)

    return op_result


@operation
def template_file(template: Template):
    template.src = f"{TEMPLATED_FILES_SUFFIX}/{template.src}"
    yield FunctionCommand(do_template, [template], {})


def run(config):
    for template in config.templates:
        template = Template(**template)
        template.dest = os.path.expanduser(template.dest)
        template_file(template)
