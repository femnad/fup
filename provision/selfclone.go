package provision

import (
	"errors"

	"github.com/femnad/fup/entity"
)

func cloneRepos(repos []entity.Repo, clonePath string) error {
	var errs []error
	for _, repo := range repos {
		path := clonePath
		if repo.Path != "" {
			path = repo.Path
		}
		err := entity.CloneUnderPath(repo, path)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func sshClone(config entity.Config) error {
	return cloneRepos(config.Repos, config.Settings.SSHCloneDir)
}
