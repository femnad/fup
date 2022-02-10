from pyinfra.operations import apt


import tasks.pkg
import tasks.config

tasks.pkg.install()
