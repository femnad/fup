package provision

import (
	"errors"
	"fmt"
	"strings"

	"github.com/femnad/fup/internal"
	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
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
		internal.Logger.Trace().Str("crate", pkg.Crate).Msg("Skipping Cargo install")
		return nil
	}

	if !when.ShouldRun(pkg) {
		internal.Logger.Trace().Str("crate", pkg.Crate).Str("when", pkg.When).Msg("Skipping Cargo install")
		return nil
	}

	name := pkg.Crate
	internal.Logger.Debug().Str("crate", pkg.Crate).Msg("Installing Cargo crate")

	installCmd := []string{"cargo", "install"}

	crate, err := crateArgs(pkg)
	if err != nil {
		internal.Logger.Error().Err(err).Str("crate", name).Msg("Error getting create name")
		return err
	}
	installCmd = append(installCmd, crate...)

	if pkg.Bins {
		installCmd = append(installCmd, "--bins")
	}

	if len(pkg.Binaries) > 0 {
		installCmd = append(installCmd, pkg.Binaries...)
	}

	cmd := strings.Join(installCmd, " ")
	_, err = run.Cmd(s, marecmd.Input{Command: cmd})
	if err != nil {
		internal.Logger.Error().Err(err).Str("name", name).Msg("Error installing cargo package")
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
