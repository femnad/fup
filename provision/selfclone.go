package provision

import (
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

const selfCloneFact = "ssh-ready"

func cloneRepos(repos []entity.Repo, clonePath string) error {
	for _, repo := range repos {
		err := common.CloneUnderPath(repo, clonePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func selfClone(config base.Config) {
	ok, err := when.FactOk(selfCloneFact)
	if err != nil {
		internal.Log.Errorf("error checking if self cloning is ok: %v", err)
		return
	}

	if !ok {
		internal.Log.Debugf("not proceeding with self cloning as fact check evaluated to fals")
		return
	}

	err = cloneRepos(config.SelfRepos, config.Settings.SelfClonePath)
	if err != nil {
		internal.Log.Errorf("error cloning own repo: %v", err)
	}
}
