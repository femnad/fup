package packages

import (
	"fmt"
	"strings"

	"github.com/femnad/mare/cmd"
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

func (Dnf) RemoveCmd() string {
	return "remove"
}

func (Dnf) RemoteInstall(urls []string) error {
	sudo, err := isUserRoot()
	if err != nil {
		return err
	}

	input := cmd.Input{Command: fmt.Sprintf("dnf install -y %s", strings.Join(urls, " ")), Sudo: sudo}
	_, err = cmd.RunFormatError(input)
	return err
}
