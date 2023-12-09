package packages

import (
	"fmt"
	"strings"

	"github.com/femnad/fup/common"
	marecmd "github.com/femnad/mare/cmd"
)

type Dnf struct {
}

func (d Dnf) ListPkgsHeader() string {
	return "Installed Packages"
}

func (Dnf) PkgExec() string {
	return "dnf"
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

func (Dnf) RemoveCmd() string {
	return "remove"
}

func (Dnf) RemoteInstall(urls []string) error {
	sudo, err := common.IsUserRoot()
	if err != nil {
		return err
	}

	input := marecmd.Input{Command: fmt.Sprintf("dnf install -y %s", strings.Join(urls, " ")), Sudo: sudo}
	_, err = marecmd.RunFormatError(input)
	return err
}
