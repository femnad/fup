import tasks.pkg
import tasks.config
import tasks.archives
import tasks.recipes
import tasks.cargo

config = tasks.config.get_config()
tasks.pkg.install(config)
tasks.archives.extract(config)

tasks.recipes.run(config)

tasks.cargo.run(config)
