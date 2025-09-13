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

		internal.Log.Infof("Adding repo %s", repo.Name())

		err := repo.Install()
		if err != nil {
			internal.Log.Errorf("Error installing repo %s: %v", repo.Name(), err)
			errs = append(errs, err)
			continue
		}
	}

	return errors.Join(errs...)
}
