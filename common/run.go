package common

import (
	"bytes"
	"fmt"
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

func RunCmdGetOutputShowError(in CmdIn) (CmdOut, error) {
	out, err := RunCmd(in)
	if err == nil {
		return out, nil
	}

	stdout := strings.TrimSpace(out.Stdout)
	stderr := strings.TrimSpace(out.Stderr)
	outStr := fmt.Sprintf("error running command %s", in.Command)

	if stdout != "" {
		outStr += fmt.Sprintf(", stdout: %s", stdout)
	}
	if stderr != "" {
		outStr += fmt.Sprintf(", stderr: %s", stderr)
	}
	outStr += fmt.Sprintf(", error: %v", err)

	return out, fmt.Errorf(outStr)
}

func RunCmdShowError(in CmdIn) error {
	_, err := RunCmdGetOutputShowError(in)
	return err
}
