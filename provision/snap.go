package provision

import (
	"errors"
	"fmt"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

func installSnap(snap entity.Snap) error {
	err := internal.Run(fmt.Sprintf("snap list %s", snap.Name))
	if err == nil {
		return nil
	}

	internal.Log.Infof("Installing snap %s", snap.Name)
	cmd := fmt.Sprintf("snap install %s", snap.Name)
	if snap.Classic {
		cmd += " --classic"
	}

	err = internal.MaybeRunWithSudo(cmd)
	if err != nil {
		internal.Log.Errorf("error installing snap %s: %v", snap.Name, err)
		return err
	}

	return nil
}

func snapInstall(config base.Config) error {
	_, err := common.Which("snap")
	if err != nil {
		internal.Log.Debug("skipping installing snap packages as snap is not available")
		return nil
	}

	var snapErr []error
	for _, snap := range config.SnapPackages {
		err = installSnap(snap)
		snapErr = append(snapErr, err)
	}

	return errors.Join(snapErr...)
}
