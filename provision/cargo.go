package provision

import (
	"errors"
	"fmt"
	"strings"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
	"github.com/femnad/fup/run"
	"github.com/femnad/fup/settings"
)

func crateArgs(name string) ([]string, error) {
	if !strings.HasPrefix(name, "https://") {
		return []string{name}, nil
	}

	crate, err := common.NameFromRepo(name)
	if err != nil {
		return nil, fmt.Errorf("error getting repo name for %s: %v", name, err)
	}

	return []string{"--git", name, crate}, nil
}

func cargoInstall(pkg entity.CargoPkg, s settings.Settings) error {
	if unless.ShouldSkip(pkg, s) {
		internal.Log.Debugf("skipping cargo install for %s", pkg.Crate)
		return nil
	}

	if !when.ShouldRun(pkg) {
		internal.Log.Debugf("skipping cargo install for %s", pkg.Crate)
		return nil
	}

	name := pkg.Crate
	internal.Log.Infof("Installing cargo package: %s", name)

	installCmd := []string{"cargo", "install"}

	crate, err := crateArgs(name)
	if err != nil {
		internal.Log.Errorf("error getting crate name for %s: %v", name, err)
		return err
	}
	installCmd = append(installCmd, crate...)

	if pkg.Bins {
		installCmd = append(installCmd, "--bins")
	}

	cmd := strings.Join(installCmd, " ")
	resp, err := run.Cmd(s, marecmd.Input{Command: cmd})
	if err != nil {
		internal.Log.Errorf("error installing cargo package %s: %v, output: %s", name, err, resp.Stderr)
		return err
	}

	return nil
}

func cargoInstallPkgs(cfg entity.Config) error {
	var cargoErr []error
	for _, pkg := range cfg.Cargo {
		err := cargoInstall(pkg, cfg.Settings)
		cargoErr = append(cargoErr, err)
	}

	return errors.Join(cargoErr...)
}
