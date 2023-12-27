package provision

import (
	"errors"
	"fmt"
	"os"
	"path"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
)

func pipInstall(pipBin, pkg string) error {
	cmd := fmt.Sprintf("%s install %s", pipBin, pkg)
	_, err := marecmd.RunFormatError(marecmd.Input{Command: cmd})
	if err != nil {
		return err
	}

	return nil
}

func pythonInstall(pkg entity.PythonPkg, cfg entity.Config) error {
	if unless.ShouldSkip(pkg, cfg.Settings) {
		internal.Log.Debugf("skipping pip install for %s", pkg.Name())
		return nil
	}

	internal.Log.Infof("Installing Python package %s", pkg.Name())

	name := pkg.Name()
	baseDir := internal.ExpandUser(cfg.Settings.VirtualEnvDir)
	venvDir := path.Join(baseDir, name)

	cmd := fmt.Sprintf("virtualenv %s", venvDir)
	_, err := marecmd.RunFormatError(marecmd.Input{Command: cmd})
	if err != nil {
		internal.Log.Errorf("error creating virtualenv for package %s: %v", name, err)
		return err
	}

	venvPip := path.Join(venvDir, "bin", "pip")

	err = pipInstall(venvPip, name)
	if err != nil {
		internal.Log.Errorf("error installing pip package %s: %v", name, err)
		return err
	}

	for _, req := range pkg.Reqs {
		err = pipInstall(venvPip, req)
		if err != nil {
			internal.Log.Errorf("error installing required pip package %s for %s: %v", req, name, err)
			return err
		}
	}

	home := os.Getenv("HOME")
	homeBin := path.Join(home, "bin")

	if len(pkg.BinLinks) == 0 {
		pkg.BinLinks = []string{pkg.Name()}
	}
	for _, link := range pkg.BinLinks {
		linkName := path.Join(homeBin, link)
		linkTarget := path.Join(venvDir, "bin", link)
		err = common.Symlink(linkName, linkTarget)
		if err != nil {
			internal.Log.Errorf("error linking from %s to %s for pkg %s: %v", linkName, linkTarget, name, err)
			return err
		}
	}

	return nil
}

func pythonInstallPkgs(cfg entity.Config) error {
	var pyErr []error
	for _, pkg := range cfg.Python {
		err := pythonInstall(pkg, cfg)
		pyErr = append(pyErr, err)
	}

	return errors.Join(pyErr...)
}
