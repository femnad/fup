import tasks.accept_host_keys
import tasks.archives
import tasks.cargo
import tasks.clone_repos
import tasks.config
import tasks.ensure_lines
import tasks.github_keys
import tasks.gopkg
import tasks.pip
import tasks.pkg
import tasks.preflight
import tasks.recipes
import tasks.services
import tasks.templates

config = tasks.config.get_config()

tasks.preflight.run(config)

tasks.pkg.run(config)

tasks.accept_host_keys.run(config)
tasks.archives.run(config)
tasks.cargo.run(config)
tasks.clone_repos.run(config)
tasks.ensure_lines.run(config)
tasks.github_keys.run(config)
tasks.gopkg.run(config)
tasks.pip.run(config)
tasks.recipes.run(config)
tasks.services.run(config)
tasks.templates.run(config)
