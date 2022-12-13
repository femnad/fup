package common

import (
	"bytes"
	"os/exec"
	"strings"
)

func RunCmd(command string) (string, error) {
	cmds := strings.Split(command, " ")
	cmd := exec.Command(cmds[0], cmds[1:]...)
	output, err := cmd.Output()
	return string(output), err
}

func RunCmdGetStderr(command string) (string, error) {
	var b bytes.Buffer
	cmds := strings.Split(command, " ")

	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Run()

	return b.String(), err
}
