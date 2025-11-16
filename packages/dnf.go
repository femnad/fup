package packages

import (
	"fmt"
	"strings"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

type Dnf struct {
}

func (Dnf) ListPkgsHeader() string {
	return "Installed packages"
}

func (Dnf) PkgExec() string {
	return "dnf"
}

func (Dnf) PkgInstallArgs() []string {
	return []string{}
}

func (Dnf) PkgEnv() map[string]string {
	return nil
}

func (Dnf) PkgNameSeparator() string {
	return "."
}

func (Dnf) PreserveEnv() bool {
	return false
}

func (Dnf) RemoveCmd() []string {
	return []string{"remove"}
}

func (Dnf) RemoteInstall(pkgs []entity.RemotePackage) error {
	var regularUrls []string
	var skipScriptUrls []string

	for _, pkg := range pkgs {
		url := pkg.Url
		if pkg.SkipScripts {
			skipScriptUrls = append(skipScriptUrls, url)
		} else {
			regularUrls = append(regularUrls, url)
		}
	}

	if len(regularUrls) > 0 {
		cmd := fmt.Sprintf("dnf install -y %s", strings.Join(regularUrls, " "))
		err := internal.MaybeRunWithSudo(cmd)
		if err != nil {
			return err
		}
	}

	if len(skipScriptUrls) > 0 {
		cmd := fmt.Sprintf("dnf install -y --setopt=tsflags=noscripts %s", strings.Join(skipScriptUrls, " "))
		err := internal.MaybeRunWithSudo(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}
