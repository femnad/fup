package provision

import (
	"fmt"
	"os"
	"path"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

func pipInstall(pipBin, pkg string) error {
	cmd := fmt.Sprintf("%s install %s", pipBin, pkg)
	_, err := common.RunCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

func pythonInstall(pkg base.PythonPkg, cfg base.Config) {
	name := pkg.Name()
	baseDir := internal.ExpandUser(cfg.Settings.VirtualEnvDir)
	venvDir := path.Join(baseDir, name)

	cmd := fmt.Sprintf("virtualenv %s", venvDir)
	_, err := common.RunCmd(cmd)
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
		err := pipInstall(venvPip, req)
		if err != nil {
			internal.Log.Errorf("error installing requirement pip package %s: %v", name, err)
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
