package internal

import (
	"fmt"
	"os"
	"path"

	marecmd "github.com/femnad/mare/cmd"
)

func maybeWarnPasswordRequired(cmdStr string) {
	out, _ := marecmd.Run(marecmd.Input{Command: "sudo -Nnv"})
	if out.Code == 0 {
		return
	}

	Log.Warningf("Sudo authentication required for escalating privileges to run command `%s`", cmdStr)
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
	err = marecmd.RunErrOnly(cmd)
	return err
}

func MaybeRunWithSudoForPath(cmdStr, targetPath string) error {
	if !path.IsAbs(targetPath) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		targetPath = path.Join(wd, targetPath)
	}

	needsSudo, err := needsSudoForPath(targetPath)
	if err != nil {
		return err
	}

	if needsSudo {
		maybeWarnPasswordRequired(cmdStr)
	}

	cmd := marecmd.Input{Command: cmdStr, Sudo: needsSudo}
	err = marecmd.RunErrOnly(cmd)
	return err
}

func Move(src, dst string, setOwner bool) error {
	mv := fmt.Sprintf("mv %s %s", src, dst)
	err := MaybeRunWithSudoForPath(mv, dst)
	if err != nil {
		return err
	}

	if !setOwner {
		return nil
	}

	err = Chown(dst, rootUser, rootUser)
	if err != nil {
		return err
	}

	return Chmod(dst, defaultFileMode)
}
