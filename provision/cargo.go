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

const (
	defaultProtocol = "https://"
)

func crateArgs(pkg entity.CargoPkg) ([]string, error) {
	name := pkg.Name()
	if !strings.Contains(name, "/") {
		return []string{name}, nil
	}
	if !strings.HasPrefix(name, defaultProtocol) {
		name = fmt.Sprintf("%s%s/%s", defaultProtocol, defaultHost, name)
	}

	crate, err := common.NameFromRepo(name)
	if err != nil {
		return nil, fmt.Errorf("error getting repo name for %s: %v", name, err)
	}

	args := []string{"--git", name, crate}
	ref := pkg.Ref
	tag := pkg.Tag
	if ref != "" {
		args = append(args, []string{"--ref", ref}...)
	} else if tag != "" {
		args = append(args, []string{"--tag", tag}...)
	}

	return args, nil
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

	crate, err := crateArgs(pkg)
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
