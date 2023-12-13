package precheck

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/internal"
)

const (
	batteryDevicePattern = "^BAT[0-9]+$"
	onepasswordSSHSocket = "~/.1password/agent.sock"
	sysClassPower        = "/sys/class/power_supply"
)

var (
	batteryDeviceRegex = regexp.MustCompile(batteryDevicePattern)
	caps               = map[string]func() (bool, error){
		"laptop": isLaptop,
		"ssh":    sshReady,
	}
	pkgMgrToOs = map[string][]string{
		"apt": {"debian", "ubuntu"},
		"dnf": {"fedora"},
	}
)

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

func sshReady() (bool, error) {
	_, err := os.Stat(internal.ExpandUser(onepasswordSSHSocket))
	if err == nil {
		return true, nil
	}

	return hasOutput("ssh-add -l")
}

func hasEnv(env string) (bool, error) {
	val := os.Getenv(env)
	return val != "", nil
}

func hasOutput(cmd string) (bool, error) {
	out, err := marecmd.Run(marecmd.Input{Command: cmd})
	if err != nil {
		return false, err
	}

	return len(strings.TrimSpace(out.Stdout)) > 0, nil
}

func hasPkgMgr(pkgMgr string) (bool, error) {
	osList, ok := pkgMgrToOs[pkgMgr]
	if !ok {
		return false, fmt.Errorf("unknown package manager: %s", pkgMgr)
	}

	for _, osName := range osList {
		res, err := isOs(osName)
		if err != nil {
			return false, err
		}
		if res {
			return true, nil
		}
	}

	return false, nil
}

func isOk(cap string) (bool, error) {
	capFn, ok := caps[cap]
	if !ok {
		return false, fmt.Errorf("no such capability check: %s", cap)
	}

	return capFn()
}

func isOs(osId string) (bool, error) {
	foundOsId, err := GetOsId()
	if err != nil {
		return false, fmt.Errorf("error getting OS ID %v", err)
	}

	return foundOsId == osId, nil
}

var FactFns = template.FuncMap{
	"env":    hasEnv,
	"pkg":    hasPkgMgr,
	"ok":     isOk,
	"os":     isOs,
	"output": hasOutput,
}
