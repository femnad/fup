package provision

import (
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

const sshReadyStatement = `ok "ssh"`

func cloneRepos(repos []entity.Repo, clonePath string) error {
	for _, repo := range repos {
		path := clonePath
		if repo.Path != "" {
			path = repo.Path
		}
		err := common.CloneUnderPath(repo, path)
		if err != nil {
			return err
		}
	}

	return nil
}

func sshClone(config base.Config) error {
	ok, err := when.EvalStatement(sshReadyStatement)
	if err != nil {
		internal.Log.Errorf("error checking if SSH cloning is ok: %v", err)
		return err
	}

	if !ok {
		internal.Log.Debugf("not proceeding with SSH cloning as fact check evaluated to false")
		return nil
	}

	err = cloneRepos(config.Repos, config.Settings.SSHCloneDir)
	if err != nil {
		internal.Log.Errorf("error SSH cloning repo: %v", err)
		return err
	}

	return nil
}
