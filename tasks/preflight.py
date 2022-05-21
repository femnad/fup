from tasks.recipes import Recipe, run_recipe


def run(config):
    for recipe in config.preflight:
        recipe = Recipe(**recipe)
        run_recipe(recipe, config.settings)
