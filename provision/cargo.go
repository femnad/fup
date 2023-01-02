package provision

import (
	"fmt"
	"strings"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/unless"
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

func cargoInstall(pkg base.CargoPkg, s base.Settings) {
	if precheck.ShouldSkip(pkg, s) {
		internal.Log.Debugf("skipping cargo install for %s", pkg.Crate)
		return
	}

	name := pkg.Crate
	internal.Log.Infof("Installing cargo package: %s", name)

	installCmd := []string{"cargo", "install"}

	crate, err := crateArgs(name)
	if err != nil {
		internal.Log.Errorf("error getting crate name for %s: %v", name, err)
		return
	}
	installCmd = append(installCmd, crate...)

	if pkg.Bins {
		installCmd = append(installCmd, "--bins")
	}

	cmd := strings.Join(installCmd, " ")
	output, err := common.RunCmdGetStderr(cmd)
	if err != nil {
		internal.Log.Errorf("error installing cargo package %s: %v, output: %s", name, err, output)
	}
}

func cargoInstallPkgs(cfg base.Config) {
	for _, pkg := range cfg.Cargo {
		cargoInstall(pkg, cfg.Settings)
	}
}
