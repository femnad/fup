import tasks.accept_host_keys
import tasks.archives
import tasks.cargo
import tasks.clone_repos
import tasks.config
import tasks.github_keys
import tasks.gopkg
import tasks.pip
import tasks.pkg
import tasks.recipes
import tasks.services
import tasks.templates

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
