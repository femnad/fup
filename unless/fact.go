package precheck

import (
	"fmt"
	"os"
	"strings"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

const (
	neovimPluginsDir = "~/.local/share/plugged"
	passwordStoreDir = "~/.password-store"
	tmuxEnvKey       = "TMUX"
)

func InTmux() (bool, error) {
	val := os.Getenv(tmuxEnvKey)
	return val != "", nil
}

func isOs(osId string) (bool, error) {
	foundOsId, err := GetOsId()
	if err != nil {
		return false, fmt.Errorf("error getting OS ID %v", err)
	}

	return foundOsId == osId, nil
}

func IsDebian() (bool, error) {
	return isOs("debian")
}

func IsFedora() (bool, error) {
	return isOs("fedora")
}

func IsUbuntu() (bool, error) {
	return isOs("ubuntu")
}

func NeovimReady() (bool, error) {
	d := internal.ExpandUser(neovimPluginsDir)
	_, err := os.Stat(d)
	return err == nil, nil
}

func SshReady() (bool, error) {
	output, err := common.RunCmd("ssh-add -l")
	if err != nil {
		return false, err
	}

	output = strings.TrimSpace(output)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Split(line, " ")
		if len(fields) != 4 {
			return false, fmt.Errorf("Unexpected SSH agent output: %s", output)
		}

		hostname, err := os.Hostname()
		if err != nil {
			return false, err
		}

		if hostname == fields[2] {
			return true, nil
		}
	}

	return false, nil
}

func SshPullReady() (bool, error) {
	d := internal.ExpandUser(passwordStoreDir)
	_, err := os.Stat(d)
	return err == nil, nil
}

var Facts = map[string]func() (bool, error){
	"in-tmux":        InTmux,
	"is-debian":      IsDebian,
	"is-fedora":      IsFedora,
	"is-ubuntu":      IsUbuntu,
	"neovim-ready":   NeovimReady,
	"ssh-pull-ready": SshPullReady,
	"ssh-ready":      SshReady,
}
