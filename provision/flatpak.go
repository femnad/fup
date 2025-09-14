package provision

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/settings"
	marecmd "github.com/femnad/mare/cmd"
)

const (
	defaultRemote  = "flathub"
	flatpakExec    = "flatpak"
	launcherScript = `#!/usr/bin/env bash
flatpak run %s $@
`
)

func ensureRemote(remote entity.FlatpakRemote) error {
	out, _ := marecmd.Run(marecmd.Input{Command: fmt.Sprintf("%s remote-ls %s", flatpakExec, remote.Name)})
	if out.Code == 0 {
		return nil
	}

	internal.Logger.Debug().Str("name", remote.Name).Msg("Adding flatpak remote")
	cmd := fmt.Sprintf("%s remote-add %s %s", flatpakExec, remote.Name, remote.Url)
	err := marecmd.RunErrOnly(marecmd.Input{Command: cmd})
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
	if pkg.Remote == "" {
		for _, remote := range remotes {
			if remote.Name != defaultRemote {
				continue
			}
			return ensureRemote(remote)
		}

		return fmt.Errorf("unable to determine remote for %s", pkg.Name)
	}

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
	out, _ := marecmd.RunFmtErr(marecmd.Input{Command: fmt.Sprintf("%s info %s", flatpakExec, pkg.Name)})
	return out.Code == 0, nil
}

func ensureInstalled(pkg entity.FlatpakPkg) error {
	out, _ := marecmd.Run(marecmd.Input{Command: fmt.Sprintf("%s info %s", pkg, pkg.Name)})
	if out.Code == 0 {
		return nil
	}

	internal.Logger.Info().Str("name", pkg.Name).Msg("Installing flatpak package")
	cmd := fmt.Sprintf("%s install %s %s -y", flatpakExec, pkg.Remote, pkg.Name)
	err := marecmd.RunErrOnly(marecmd.Input{Command: cmd})
	if err != nil {
		return fmt.Errorf("error installing flatpak %s: %v", pkg.Name, err)
	}

	return nil
}

func ensureLauncher(stg settings.Settings, flatpak entity.FlatpakPkg) error {
	if flatpak.Launcher == "" {
		return nil
	}

	homeBin := internal.ExpandUser(stg.BinDir)
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

func installFlatpak(stg settings.Settings, pkg entity.FlatpakPkg, remotes []entity.FlatpakRemote) error {
	installed, err := isInstalled(pkg)
	if err != nil {
		return err
	}

	if !installed {
		err = ensurePkgRemote(pkg, remotes)
		if err != nil {
			return err
		}

		err = ensureInstalled(pkg)
		if err != nil {
			return err
		}
	}

	return ensureLauncher(stg, pkg)
}

func flatpakInstall(config entity.Config) error {
	_, err := common.Which(flatpakExec)
	if err != nil {
		internal.Logger.Warn().Msg("Flatpak not installed, skipping installation")
		return nil
	}

	var flatpakErr []error
	for _, pkg := range config.Flatpak.Packages {
		err = installFlatpak(config.Settings, pkg, config.Flatpak.Remotes)
		if err != nil {
			internal.Logger.Error().Err(err).Str("package", pkg.Name).Msg("Error installing Flatpak")
		}
		flatpakErr = append(flatpakErr, err)
	}

	return errors.Join(flatpakErr...)
}
