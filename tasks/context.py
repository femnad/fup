from typing import Dict


def expand(s: str, context: Dict[str, str]):
    cur_dlr_index = -1
    parsing_var = False

    varmap = {}
    cur_var = ''

    for i, c in enumerate(s):
        if c == '$':
            cur_dlr_index = i
            continue
        elif c == '{' and i == cur_dlr_index + 1:
            parsing_var = True
            continue
        elif c == '}':
            parsing_var = False
            value = context[cur_var]
            varmap[cur_var] = value
            cur_var = ''
            cur_dlr_index = -1
        elif parsing_var:
            cur_var += c

    for var, val in varmap.items():
        s = s.replace(f'${{{var}}}', str(val))

    return s
