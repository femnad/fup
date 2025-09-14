package packages

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
)

const (
	dpkg          = "dpkg"
	pkgScriptsDir = "/var/lib/dpkg/info"
)

var pkgScripts = []string{"postinst", "preinst"}

type Apt struct {
}

func (Apt) ListPkgsHeader() string {
	return "Listing..."
}

func (Apt) PkgExec() string {
	return "apt"
}

func (Apt) PkgInstallArgs() []string {
	return []string{"-U"}
}

func (Apt) PkgEnv() map[string]string {
	return map[string]string{
		"DEBIAN_FRONTEND": "noninteractive",
		"DEBIAN_PRIORITY": "critical",
	}
}

func (Apt) PreserveEnv() bool {
	return true
}

func (Apt) PkgNameSeparator() string {
	return "/"
}

func (Apt) RemoveCmd() []string {
	return []string{"purge", "--auto-remove"}
}

func (Apt) UpdateCmd() string {
	return "update"
}

func installPkgSkipScripts(pkgName, filename string) error {
	err := internal.MaybeRunWithSudo(fmt.Sprintf("%s --unpack %s", dpkg, filename))
	if err != nil {
		return err
	}

	for _, script := range pkgScripts {
		scriptFile := path.Join(pkgScriptsDir, fmt.Sprintf("%s.%s", pkgName, script))
		err = internal.MaybeRunWithSudo(fmt.Sprintf("rm %s", scriptFile))
		if err != nil {
			return err
		}
	}

	return internal.MaybeRunWithSudo(fmt.Sprintf("%s --configure %s", dpkg, pkgName))
}

func (Apt) RemoteInstall(pkgs []entity.RemotePackage) error {
	tmpDir, err := os.MkdirTemp("/tmp", "fup-remote-pkg")
	if err != nil {
		return err
	}

	skipScriptTargets := make(map[string]string)
	var regularTargets []string

	for _, pkg := range pkgs {
		url := pkg.Url
		_, file := path.Split(url)
		target := path.Join(tmpDir, file)
		err = remote.Download(url, target)
		if err != nil {
			return err
		}

		if pkg.SkipScripts {
			skipScriptTargets[pkg.Name] = target
		} else {
			regularTargets = append(regularTargets, target)
		}
	}

	targetArgs := strings.Join(regularTargets, " ")
	if targetArgs != "" {
		err = internal.MaybeRunWithSudo(fmt.Sprintf("apt install -Uy %s", targetArgs))
		if err != nil {
			return err
		}
	}

	for pkg, filename := range skipScriptTargets {
		err = installPkgSkipScripts(pkg, filename)
		if err != nil {
			return err
		}
	}

	return os.RemoveAll(tmpDir)
}
