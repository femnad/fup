package provision

import (
	"errors"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
)

func addRepos(config entity.Config) error {
	var errs []error

	var repos []entity.OSRepo
	for _, repo := range config.AptRepos {
		repos = append(repos, repo)
	}
	for _, repo := range config.DnfRepos {
		repos = append(repos, repo)
	}

	for _, repo := range repos {
		if !when.ShouldRun(repo) {
			continue
		}

		if unless.ShouldSkip(repo, config.Settings) {
			continue
		}

		exists, err := repo.Exists()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if exists {
			continue
		}
		internal.Logger.Debug().Str("name", repo.Name()).Msg("Adding repo")

		err = repo.Install()
		if err != nil {
			internal.Logger.Error().Err(err).Str("name", repo.Name()).Msg("Error installing repo")
			errs = append(errs, err)
			continue
		}
	}

	return errors.Join(errs...)
}
