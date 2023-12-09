package provision

import (
	"errors"
	"fmt"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	marecmd "github.com/femnad/mare/cmd"
)

func installSnap(snap entity.Snap) error {
	cmd := fmt.Sprintf("snap install %s", snap.Name)
	if snap.Classic {
		cmd += " --classic"
	}
	in := marecmd.Input{Command: cmd}

	_, err := marecmd.RunFormatError(in)
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
