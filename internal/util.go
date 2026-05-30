package internal

import (
	"fmt"
	"strings"

	marecmd "github.com/femnad/mare/cmd"
)

func PipInstall(pipBin, pkg, version string, user bool) error {
	cmd := fmt.Sprintf("%s install %s", pipBin, pkg)
	if version != "" {
		cmd += fmt.Sprintf("==%s", version)
	}
	if user {
		cmd += fmt.Sprintf(" --user")
	}

	err := marecmd.RunErrOnly(marecmd.Input{Command: cmd})
	if err != nil {
		return err
	}

	return nil
}

func PrettyLogStr(field string) string {
	return strings.Replace(field, "\"", "`", -1)
}
