package internal

import (
	"strings"

	marecmd "github.com/femnad/mare/cmd"
)

func maybeWarnPasswordRequired(cmdStr string) {
	out, _ := marecmd.Run(marecmd.Input{Command: "sudo -Nnv"})
	if out.Code == 0 {
		return
	}

	cmdHead := strings.Split(cmdStr, " ")[0]
	Log.Warningf("Sudo authentication required for escalating privileges to run command %s", cmdHead)
}

func needsSudoForPath(dst string) (bool, error) {
	isRoot, err := IsUserRoot()
	if err != nil {
		return false, err
	}

	var sudo bool
	if isRoot {
		sudo = false
	} else {
		sudo = !IsHomePath(dst)
	}

	return sudo, nil
}

func MaybeRunWithSudo(cmdStr string) error {
	isRoot, err := IsUserRoot()
	if err != nil {
		return err
	}

	if !isRoot {
		maybeWarnPasswordRequired(cmdStr)
	}

	cmd := marecmd.Input{Command: cmdStr, Sudo: !isRoot}
	_, err = marecmd.RunFormatError(cmd)
	return err
}

func MaybeRunWithSudoForPath(cmdStr, path string) error {
	needsSudo, err := needsSudoForPath(path)
	if err != nil {
		return err
	}

	if needsSudo {
		maybeWarnPasswordRequired(cmdStr)
	}

	cmd := marecmd.Input{Command: cmdStr, Sudo: needsSudo}
	_, err = marecmd.RunFormatError(cmd)
	return err
}
