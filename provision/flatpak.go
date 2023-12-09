package provision

import (
	"errors"
	"fmt"
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	marecmd "github.com/femnad/mare/cmd"
	"os"
	"path"
)

const (
	launcherScript = `#!/usr/bin/env bash
flatpak run %s
`
)

var (
	steps = []func(f entity.FlatpakPkg) error{ensureInstalled, ensureLauncher}
)

func ensureRemote(remote entity.FlatpakRemote) error {
	out, _ := marecmd.Run(marecmd.Input{Command: fmt.Sprintf("flatpak remote-ls %s", remote.Name)})
	if out.Code == 0 {
		return nil
	}

	internal.Log.Debugf("Adding flatpak remote %s", remote.Name)
	_, err := marecmd.RunFormatError(marecmd.Input{
		Command: fmt.Sprintf("flatpak remote-add %s %s", remote.Name, remote.Url)})
	if err != nil {
		return fmt.Errorf("error adding flatpak remote %s with URL %s: %v", remote.Name, remote.Url, err)
	}

	return nil
}

func ensureRemotes(remotes []entity.FlatpakRemote) error {
	for _, remote := range remotes {
		err := ensureRemote(remote)
		if err != nil {
			internal.Log.Error(err)
		}
		return err
	}

	return nil
}

func ensureInstalled(flatpak entity.FlatpakPkg) error {
	out, _ := marecmd.Run(marecmd.Input{Command: fmt.Sprintf("flatpak info %s", flatpak.Name)})
	if out.Code == 0 {
		return nil
	}

	internal.Log.Debug("Installing flatpak package %s", flatpak.Name)
	_, err := marecmd.RunFormatError(marecmd.Input{
		Command: fmt.Sprintf("flatpak install -y %s", flatpak.Name)})
	if err != nil {
		return fmt.Errorf("error install flatpak %s: %v", flatpak.Name, err)
	}

	return nil
}

func ensureLauncher(flatpak entity.FlatpakPkg) error {
	if flatpak.Launcher == "" {
		return nil
	}

	home := os.Getenv("HOME")
	homeBin := path.Join(home, "bin")
	launcherPath := path.Join(homeBin, flatpak.Launcher)

	_, err := os.Stat(launcherPath)
	if err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	launcherContent := fmt.Sprintf(launcherScript, flatpak.Name)
	return os.WriteFile(launcherPath, []byte(launcherContent), 0755)
}

func installFlatpak(flatpak entity.FlatpakPkg) error {
	for _, step := range steps {
		err := step(flatpak)
		if err != nil {
			internal.Log.Error(err)
			return err
		}
	}

	return nil
}

func flatpakInstall(config base.Config) error {
	err := ensureRemotes(config.Flatpak.Remotes)
	if err != nil {
		return err
	}

	var flatpakErr []error
	for _, flatpak := range config.Flatpak.Packages {
		err = installFlatpak(flatpak)
		flatpakErr = append(flatpakErr, err)
	}

	return errors.Join(flatpakErr...)
}
