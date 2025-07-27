package provision

import (
	"errors"
	"fmt"
	"path"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
)

func pipInstall(pipBin, pkg, version string) error {
	cmd := fmt.Sprintf("%s install %s", pipBin, pkg)
	if version != "" {
		cmd += fmt.Sprintf("==%s", version)
	}

	err := marecmd.RunErrOnly(marecmd.Input{Command: cmd})
	if err != nil {
		return err
	}

	return nil
}

func pythonInstall(pkg entity.PythonPkg, cfg entity.Config) error {
	name := pkg.Name()
	baseDir := internal.ExpandUser(cfg.Settings.VirtualEnvDir)
	venvDir := path.Join(baseDir, name)
	venvPip := path.Join(venvDir, "bin", "pip")

	if pkg.Library {
		pkg.Unless = unless.Unless{
			Cmd: fmt.Sprintf("%s show %s", venvPip, name),
		}
		if pkg.GetVersion() != "" {
			pkg.Unless.Post = `head 1 | splitBy ": " -1`
		}
	}

	if unless.ShouldSkip(pkg, cfg.Settings) {
		internal.Log.Debugf("skipping pip install for %s", pkg.Name())
		return nil
	}

	internal.Log.Infof("Installing Python package %s", pkg.Name())

	cmd := fmt.Sprintf("virtualenv %s", venvDir)
	err := marecmd.RunErrOnly(marecmd.Input{Command: cmd})
	if err != nil {
		internal.Log.Errorf("error creating virtualenv for package %s: %v", name, err)
		return err
	}

	version, err := pkg.LookupVersion(cfg.Settings)
	if err != nil {
		return err
	}

	err = pipInstall(venvPip, name, version)
	if err != nil {
		internal.Log.Errorf("error installing pip package %s: %v", name, err)
		return err
	}

	for _, req := range pkg.Reqs {
		err = pipInstall(venvPip, req, "")
		if err != nil {
			internal.Log.Errorf("error installing required pip package %s for %s: %v", req, name, err)
			return err
		}
	}

	homeBin := internal.ExpandUser(cfg.Settings.BinDir)
	if len(pkg.BinLinks) == 0 && !pkg.Library {
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
