package provision

import (
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

func cloneRepos(repos []entity.Repo, clonePath string) error {
	for _, repo := range repos {
		err := common.CloneRepo(repo, clonePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func selfClone(config base.Config) {
	err := cloneRepos(config.SelfRepos, config.Settings.SelfClonePath)
	if err != nil {
		internal.Log.Errorf("error cloning own repo: %v", err)
	}
}
