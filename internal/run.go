package internal

import marecmd "github.com/femnad/mare/cmd"

func MaybeRunWithSudo(cmdStr string) error {
	isRoot, err := IsUserRoot()
	if err != nil {
		return err
	}

	cmd := marecmd.Input{Command: cmdStr, Sudo: !isRoot}
	_, err = marecmd.RunFormatError(cmd)
	return err
}

func Run(cmdStr string) error {
	cmd := marecmd.Input{Command: cmdStr}
	_, err := marecmd.RunFormatError(cmd)
	return err
}
