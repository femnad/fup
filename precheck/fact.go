package precheck

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/internal"
)

const (
	batteryDevicePattern = "^BAT[0-9]+$"
	gcloudCredentials    = "~/.local/share/password-store"
	gopathEnvKey         = "GOPATH"
	neovimPluginsDir     = "~/.local/share/plugged"
	sysClassPower        = "/sys/class/power_supply"
	tmuxEnvKey           = "TMUX"
)

var (
	batteryDeviceRegex = regexp.MustCompile(batteryDevicePattern)
)

func goPathSet() (bool, error) {
	return os.Getenv(gopathEnvKey) != "", nil
}

func isLaptop() (bool, error) {
	_, err := os.Stat(sysClassPower)
	if err != nil {
		return false, err
	}

	entries, err := os.ReadDir(sysClassPower)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if batteryDeviceRegex.MatchString(entry.Name()) {
			return true, nil
		}
	}

	return false, nil
}

func inTmux() (bool, error) {
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

func isDebian() (bool, error) {
	return isOs("debian")
}

func isFedora() (bool, error) {
	return isOs("fedora")
}

func isUbuntu() (bool, error) {
	return isOs("ubuntu")
}

func neovimReady() (bool, error) {
	d := internal.ExpandUser(neovimPluginsDir)
	_, err := os.Stat(d)
	return err == nil, nil
}

func sshReady() (bool, error) {
	resp, _ := marecmd.RunFormatError(marecmd.Input{Command: "ssh-add -l"})
	if resp.Code == 1 {
		return false, nil
	}

	output := strings.TrimSpace(resp.Stdout)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Split(line, " ")
		if len(fields) != 4 {
			return false, fmt.Errorf("unexpected SSH agent output: %s", output)
		}

		hostname, err := os.Hostname()
		if err != nil {
			return false, err
		}

		// Third field could be <user>@<hostname> or <hostname>.
		if strings.HasSuffix(fields[2], hostname) {
			return true, nil
		}
	}

	return false, nil
}

func sshPullReady() (bool, error) {
	d := internal.ExpandUser(gcloudCredentials)
	_, err := os.Stat(d)
	return err == nil, nil
}

var Facts = map[string]func() (bool, error){
	"gopath-set":     goPathSet,
	"is-laptop":      isLaptop,
	"in-tmux":        inTmux,
	"is-debian":      isDebian,
	"is-fedora":      isFedora,
	"is-ubuntu":      isUbuntu,
	"neovim-ready":   neovimReady,
	"ssh-pull-ready": sshPullReady,
	"ssh-ready":      sshReady,
}
