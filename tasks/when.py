from pyinfra import host

import facts.base


def should_run(when):
    if not when:
        return True

    negate = False
    if ' ' in when and when.startswith('not '):
        negate = True
        when = when.split()[-1]

    fact_class = ''.join([w.capitalize() for w in when.split('-')])
    result = host.get_fact(getattr(facts.base, fact_class))

    return not result if negate else result
