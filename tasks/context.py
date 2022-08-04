import copy
import os
import socket
from typing import Dict

import tasks.config

DEFAULT_KEY = '*'


def expand(s: str, context: Dict[str, str] = {}):
    updated_context = copy.deepcopy(context)

    config = tasks.config.get_config()

    home = os.getenv('HOME')
    s = s.replace('~/', f'{home}/')

    cur_dlr_index = -1
    parsing_var = False

    varmap = {}
    cur_var = ''

    lookups = []

    hostname = socket.gethostname()

    for k, v in config.settings.__dict__.items():
        if type(v) not in [int, float, str]:
            continue
        updated_context[k] = os.path.expanduser(v)

    for fact, host_pairs in config.settings.host_facts.items():
        default = host_pairs[DEFAULT_KEY]
        updated_context[fact] = host_pairs.get(hostname, default)

    for i, c in enumerate(s):
        if c == '$':
            cur_dlr_index = i
            continue
        elif c == '{' and i == cur_dlr_index + 1:
            parsing_var = True
            continue
        elif c == '}':
            parsing_var = False
            lookups.append(cur_var)
            cur_var = ''
            cur_dlr_index = -1
        elif parsing_var:
            cur_var += c

    for lookup in lookups:
        if not lookup:
            continue

        if lookup in updated_context:
            value = updated_context[lookup]
            varmap[lookup] = value
        elif value := os.getenv(lookup):
            varmap[lookup] = value
        else:
            raise Exception(f'Cannot determine value of variable `{lookup}`')

    for var, val in varmap.items():
        s = s.replace(f'${{{var}}}', str(val))

    return s
