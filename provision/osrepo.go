package provision

import (
	"errors"
	"log/slog"

	"github.com/femnad/fup/entity"
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
		slog.Info("Adding repo", "name", repo.Name())

		err = repo.Install()
		if err != nil {
			slog.Error("Error installing repo", "name", repo.Name(), "error", err)
			errs = append(errs, err)
			continue
		}
	}

	return errors.Join(errs...)
}
