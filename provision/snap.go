package provision

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	marecmd "github.com/femnad/mare/cmd"
)

func isSnapInstalled(snap entity.Snap) bool {
	out, _ := marecmd.RunFmtErr(marecmd.Input{Command: fmt.Sprintf("snap list %s", snap.Name)})
	return out.Code == 0
}

func installSnap(snap entity.Snap) error {
	if isSnapInstalled(snap) {
		return nil
	}

	slog.Info("Installing snap", "name", snap.Name)
	cmd := fmt.Sprintf("snap install %s", snap.Name)
	if snap.Classic {
		cmd += " --classic"
	}

	err := internal.MaybeRunWithSudo(cmd)
	if err != nil {
		slog.Error("error installing snap", "name", snap.Name, "error", err)
		return err
	}

	return nil
}

func uninstallSnap(snap entity.Snap) error {
	if !isSnapInstalled(snap) {
		return nil
	}

	slog.Info("Uninstalling snap", "name", snap.Name)
	cmd := fmt.Sprintf("snap remove %s", snap.Name)

	err := internal.MaybeRunWithSudo(cmd)
	if err != nil {
		slog.Error("error uninstalling snap", "name", snap.Name, "error", err)
		return err
	}

	return nil
}

func snapInstall(config entity.Config) error {
	_, err := common.Which("snap")
	if err != nil {
		slog.Debug("skipping installing snap packages as snap is not available")
		return nil
	}

	var snapErr []error
	for _, snap := range config.SnapPackages {
		if snap.Absent {
			err = uninstallSnap(snap)
		} else {
			err = installSnap(snap)
		}
		snapErr = append(snapErr, err)
	}

	return errors.Join(snapErr...)
}
