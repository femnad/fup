import tasks.pkg
import tasks.config
import tasks.archives
import tasks.recipes

config = tasks.config.get_config()
tasks.pkg.install(config)
tasks.archives.extract(config)

tasks.recipes.run(config)
