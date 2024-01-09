package provision

import (
	"errors"
	"github.com/femnad/fup/settings"

	"github.com/femnad/fup/entity"
)

func cloneRepos(repos []entity.Repo, s settings.Settings) error {
	var errs []error
	for _, repo := range repos {
		path := s.SSHCloneDir
		if repo.Path != "" {
			path = repo.Path
		}
		err := entity.CloneUnderPath(repo, path, s.CloneEnv)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func sshClone(config entity.Config) error {
	return cloneRepos(config.Repos, config.Settings)
}
