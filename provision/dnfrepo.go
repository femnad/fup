package provision

import (
	"errors"
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
)

func addDnfRepos(config base.Config, repos []entity.DnfRepo) error {
	var errs []error
	for _, repo := range repos {
		if !when.ShouldRun(repo) {
			continue
		}

		if unless.ShouldSkip(repo, config.Settings) {
			continue
		}

		err := repo.Install()
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
