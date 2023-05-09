package common

import (
	"bytes"
	"os/exec"
	"strings"
)

var defaultShell = "sh"

type CmdIn struct {
	Command string
	Pwd     string
	Shell   bool
	Sudo    bool
}

type CmdOut struct {
	Code   int
	Stdout string
	Stderr string
}

func RunCmd(in CmdIn) (CmdOut, error) {
	var cmdSlice []string
	if in.Shell {
		cmdSlice = append([]string{defaultShell, "-c"}, in.Command)
	} else {
		cmdSlice = strings.Split(in.Command, " ")
	}
	if in.Sudo {
		cmdSlice = append([]string{"sudo"}, cmdSlice...)
	}

	cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)
	if in.Pwd != "" {
		cmd.Dir = in.Pwd
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return CmdOut{Stdout: stdout.String(), Stderr: stderr.String(), Code: cmd.ProcessState.ExitCode()}, err
}
