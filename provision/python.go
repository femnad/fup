package provision

import (
	"fmt"
	"os"
	"path"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
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

func pythonInstall(pkg base.PythonPkg, cfg base.Config) {
	if unless.ShouldSkip(pkg, cfg.Settings) {
		internal.Log.Debugf("skipping pip install for %s", pkg.Name())
		return
	}

	internal.Log.Infof("Installing Python package %s", pkg.Name())

	name := pkg.Name()
	baseDir := internal.ExpandUser(cfg.Settings.VirtualEnvDir)
	venvDir := path.Join(baseDir, name)

	cmd := fmt.Sprintf("virtualenv %s", venvDir)
	_, err := marecmd.RunFormatError(marecmd.Input{Command: cmd})
	if err != nil {
		internal.Log.Errorf("error creating virtualenv for package %s: %v", name, err)
		return
	}

	venvPip := path.Join(venvDir, "bin", "pip")

	err = pipInstall(venvPip, name)
	if err != nil {
		internal.Log.Errorf("error installing pip package %s: %v", name, err)
		return
	}

	for _, req := range pkg.Reqs {
		err = pipInstall(venvPip, req)
		if err != nil {
			internal.Log.Errorf("error installing required pip package %s for %s: %v", req, name, err)
			return
		}
	}

	home := os.Getenv("HOME")
	homeBin := path.Join(home, "bin")
	for _, link := range pkg.BinLinks {
		linkName := path.Join(homeBin, link)
		linkTarget := path.Join(venvDir, "bin", link)
		err = common.Symlink(linkName, linkTarget)
		if err != nil {
			internal.Log.Errorf("error linking from %s to %s for pkg %s: %v", linkName, linkTarget, name, err)
			return
		}
	}
}

func pythonInstallPkgs(cfg base.Config) {
	for _, pkg := range cfg.Python {
		pythonInstall(pkg, cfg)
	}
}
