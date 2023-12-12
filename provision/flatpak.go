package provision

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	marecmd "github.com/femnad/mare/cmd"
)

const (
	launcherScript = `#!/usr/bin/env bash
flatpak run %s
`
)

func ensureRemote(remote entity.FlatpakRemote) error {
	out, err := marecmd.Run(marecmd.Input{Command: fmt.Sprintf("flatpak remote-ls %s", remote.Name)})
	if err != nil {
		return err
	}
	if out.Code == 0 {
		return nil
	}

	internal.Log.Debugf("Adding flatpak remote %s", remote.Name)
	cmd := fmt.Sprintf("flatpak remote-add %s %s", remote.Name, remote.Url)
	_, err = marecmd.RunFormatError(marecmd.Input{Command: cmd})
	if err != nil {
		return fmt.Errorf("error adding flatpak remote %s with URL %s: %v", remote.Name, remote.Url, err)
	}

	return nil
}

func findRequiredRemote(pkg entity.FlatpakPkg, remotes []entity.FlatpakRemote) (entity.FlatpakRemote, error) {
	for _, remote := range remotes {
		if pkg.Remote == remote.Name {
			return remote, nil
		}
	}

	return entity.FlatpakRemote{}, fmt.Errorf("no Flatpak remote definition found for %s", pkg.Remote)
}

func ensurePkgRemote(pkg entity.FlatpakPkg, remotes []entity.FlatpakRemote) error {
	remote, err := findRequiredRemote(pkg, remotes)
	if err != nil {
		return err
	}

	err = ensureRemote(remote)
	if err != nil {
		return err
	}

	return nil
}

func isInstalled(pkg entity.FlatpakPkg) (bool, error) {
	out, err := marecmd.Run(marecmd.Input{Command: fmt.Sprintf("flatpak info %s", pkg.Name)})
	if err != nil {
		return false, err
	}

	return out.Code == 0, nil
}

func ensureInstalled(flatpak entity.FlatpakPkg) error {
	out, _ := marecmd.Run(marecmd.Input{Command: fmt.Sprintf("flatpak info %s", flatpak.Name)})
	if out.Code == 0 {
		return nil
	}

	internal.Log.Debug("Installing flatpak package %s", flatpak.Name)
	cmd := fmt.Sprintf("flatpak install %s %s -y", flatpak.Remote, flatpak.Name)
	_, err := marecmd.RunFormatError(marecmd.Input{Command: cmd})
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

func installFlatpak(pkg entity.FlatpakPkg, remotes []entity.FlatpakRemote) error {
	installed, err := isInstalled(pkg)
	if err != nil {
		return err
	}

	if installed {
		return nil
	}

	err = ensurePkgRemote(pkg, remotes)
	if err != nil {
		return err
	}

	err = ensureInstalled(pkg)
	if err != nil {
		return err
	}

	return ensureLauncher(pkg)
}

func flatpakInstall(config base.Config) error {
	var flatpakErr []error
	for _, flatpak := range config.Flatpak.Packages {
		err := installFlatpak(flatpak, config.Flatpak.Remotes)
		flatpakErr = append(flatpakErr, err)
	}

	return errors.Join(flatpakErr...)
}
