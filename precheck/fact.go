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
	matchers           = map[string]func() (bool, error){
		"laptop": isLaptop,
	}
	pkgMgrToOs = map[string][]string{
		"apt": {"debian", "ubuntu"},
		"dnf": {"fedora"},
	}
	readiness = map[string]func() (bool, error){
		"ssh": sshReady,
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
		return false, nil
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

func isA(matcher string) (bool, error) {
	matchFn, ok := matchers[matcher]
	if !ok {
		return false, fmt.Errorf("no such matcher: %s", matcher)
	}

	return matchFn()
}

func isOk(cap string) (bool, error) {
	capFn, ok := readiness[cap]
	if !ok {
		return false, fmt.Errorf("no such capability check: %s", cap)
	}

	return capFn()
}

func isOs(osId string) (bool, error) {
	foundOsId, err := GetOSId()
	if err != nil {
		return false, fmt.Errorf("error getting OS ID %v", err)
	}

	return foundOsId == osId, nil
}

func isOsVersion(version float64) (bool, error) {
	osVersion, err := GetOSVersion()
	if err != nil {
		return false, err
	}

	return osVersion == version, err
}

func osVersionGe(version float64) (bool, error) {
	osVersion, err := GetOSVersion()
	if err != nil {
		return false, err
	}

	return osVersion >= version, nil
}

func osVersionGt(version float64) (bool, error) {
	osVersion, err := GetOSVersion()
	if err != nil {
		return false, err
	}

	return osVersion > version, nil
}

func osVersionLe(version float64) (bool, error) {
	osVersion, err := GetOSVersion()
	if err != nil {
		return false, err
	}

	return osVersion <= version, nil
}

func osVersionLt(version float64) (bool, error) {
	osVersion, err := GetOSVersion()
	if err != nil {
		return false, err
	}

	return osVersion < version, nil
}

func statOk(path string) (bool, error) {
	path = os.ExpandEnv(internal.ExpandUser(path))
	_, err := os.Stat(path)
	return err == nil, nil
}

func negate(fn func(string) (bool, error)) func(string) (bool, error) {
	return func(in string) (bool, error) {
		out, err := fn(in)
		if err != nil {
			return false, err
		}
		return !out, nil
	}
}

var FactFns = template.FuncMap{
	"env":       hasEnv,
	"is":        isA,
	"notOs":     negate(isOs),
	"ok":        isOk,
	"os":        isOs,
	"output":    hasOutput,
	"pkg":       hasPkgMgr,
	"stat":      statOk,
	"version":   isOsVersion,
	"versionGe": osVersionGe,
	"versionGt": osVersionGt,
	"versionLe": osVersionLe,
	"versionLt": osVersionLt,
}
