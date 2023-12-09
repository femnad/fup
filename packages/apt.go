package packages

import (
	"fmt"
	"os"
	"path"
	"strings"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/remote"
)

type Apt struct {
}

func (a Apt) ListPkgsHeader() string {
	return "Listing..."
}

func (Apt) PkgExec() string {
	return "apt"
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

func (Apt) RemoveCmd() string {
	return "purge"
}

func (Apt) remoteInstall(url string) error {
	_, file := path.Split(url)
	tmpDir, err := os.MkdirTemp("/tmp", "fup-remote-pkg")
	if err != nil {
		return err
	}

	target := path.Join(tmpDir, file)
	err = remote.Download(url, target)
	if err != nil {
		return err
	}

	sudo, err := common.IsUserRoot()
	if err != nil {
		return err
	}

	input := marecmd.Input{Command: fmt.Sprintf("apt install -y %s", target), Sudo: sudo}
	return marecmd.RunNoOutput(input)
}

func (Apt) RemoteInstall(urls []string) error {
	var targets []string
	tmpDir, err := os.MkdirTemp("/tmp", "fup-remote-pkg")
	if err != nil {
		return err
	}

	for _, url := range urls {
		_, file := path.Split(url)
		target := path.Join(tmpDir, file)
		err = remote.Download(url, target)
		if err != nil {
			return err
		}

		targets = append(targets, target)
	}

	sudo, err := common.IsUserRoot()
	if err != nil {
		return err
	}

	targetArgs := strings.Join(targets, " ")
	input := marecmd.Input{Command: fmt.Sprintf("apt install -y %s", targetArgs), Sudo: sudo}
	_, err = marecmd.RunFormatError(input)
	if err != nil {
		return err
	}

	return os.RemoveAll(tmpDir)
}
