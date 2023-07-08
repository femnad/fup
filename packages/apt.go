package packages

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/femnad/mare/cmd"

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
	return map[string]string{"DEBIAN_FRONTEND": "noninteractive"}
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

	sudo, err := isUserRoot()
	if err != nil {
		return err
	}

	input := cmd.Input{Command: fmt.Sprintf("apt install -y %s", target), Sudo: sudo}
	return cmd.RunNoOutput(input)
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

	sudo, err := isUserRoot()
	if err != nil {
		return err
	}

	targetArgs := strings.Join(targets, " ")
	input := cmd.Input{Command: fmt.Sprintf("apt install -y %s", targetArgs), Sudo: sudo}
	_, err = cmd.RunFormatError(input)
	if err != nil {
		return err
	}

	return os.RemoveAll(tmpDir)
}
