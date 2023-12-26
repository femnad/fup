package internal

import marecmd "github.com/femnad/mare/cmd"

func maybeWarnPasswordRequired(cmdStr string) {
	out, _ := marecmd.Run(marecmd.Input{Command: "sudo -Nnv"})
	if out.Code == 0 {
		return
	}

	Log.Warningf("Sudo authentication required for escalating privileges to run command %s", cmdStr)
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

func Run(cmdStr string) error {
	cmd := marecmd.Input{Command: cmdStr}
	_, err := marecmd.RunFormatError(cmd)
	return err
}
