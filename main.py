import tasks.pkg
import tasks.config
import tasks.archives
import tasks.recipes
import tasks.cargo
import tasks.gopkg
import tasks.templates
import tasks.services
import tasks.github_keys
import tasks.accept_host_keys
import tasks.clone_repos

config = tasks.config.get_config()

tasks.pkg.install(config)
tasks.archives.extract(config)
tasks.github_keys.run(config)
tasks.accept_host_keys.run(config)
tasks.recipes.run(config)
tasks.cargo.run(config)
tasks.gopkg.run(config)
tasks.templates.run(config)
tasks.services.run(config)
tasks.clone_repos.run(config)
