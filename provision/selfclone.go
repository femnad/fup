package provision

import (
	"errors"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/settings"
)

func cloneRepos(repos []entity.Repo, s settings.Settings) error {
	var errs []error
	for _, repo := range repos {
		err := entity.CloneUnderPath(repo, repo.Path, s)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func sshClone(config entity.Config) error {
	return cloneRepos(config.Repos, config.Settings)
}
