package provision

import (
	"errors"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

func cloneRepos(repos []entity.Repo, clonePath string) error {
	var errs []error
	for _, repo := range repos {
		path := clonePath
		if repo.Path != "" {
			path = repo.Path
		}
		err := common.CloneUnderPath(repo, path)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func sshClone(config base.Config) error {
	for _, group := range config.RepoGroups {
		if !when.ShouldRun(group) {
			continue
		}

		err := cloneRepos(group.Clones, config.Settings.SSHCloneDir)
		if err != nil {
			internal.Log.Errorf("error cloning repos: %v", err)
			return err
		}
	}

	return nil
}
