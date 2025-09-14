package provision

import (
	"errors"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/settings"
)

func cloneRepos(repos []entity.Repo, s settings.Settings) error {
	var errs []error
	for _, repo := range repos {
		internal.Logger.Debug().Str("name", repo.Name).Msg("Cloning repo via SSH")
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
