import tasks.pkg
import tasks.config
import tasks.archives

config = tasks.config.get_config()
tasks.pkg.install(config)
tasks.archives.extract(config)
