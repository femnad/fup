import os.path

from pyinfra.operations import files

import tasks.config


def run(config: tasks.config.Config):
    for directory in config.unwanted_dirs:
        directory = os.path.expanduser(directory)
        files.directory(directory, present=False)
