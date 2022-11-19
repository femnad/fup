package provision

import (
	"fmt"
	"strings"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/unless"
)

func crateArgs(name string, multiBin bool) ([]string, error) {
	if !strings.HasPrefix(name, "https://") {
		return []string{name}, nil
	}

	args := []string{"--git", name}

	if !multiBin {
		return args, nil
	}

	repoName, err := common.NameFromRepo(name)
	if err != nil {
		return nil, fmt.Errorf("error getting repo name for %s: %v", name, err)
	}
	args = append(args, repoName)

	return args, nil
}

func cargoInstall(pkg base.CargoPkg) {
	if precheck.ShouldSkip(pkg) {
		internal.Log.Debugf("skipping cargo install for %s", pkg.Name)
		return
	}

	name := pkg.Name
	internal.Log.Infof("Installing cargo package: %s", name)

	installCmd := []string{"cargo", "install"}

	crate, err := crateArgs(name, pkg.MultiBin)
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

func cargoInstallPkgs(pkgs []base.CargoPkg) {
	for _, pkg := range pkgs {
		cargoInstall(pkg)
	}
}
