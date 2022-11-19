package common

import (
	"os/exec"
	"strings"
)

func RunCmd(command string) (string, error) {
	cmds := strings.Split(command, " ")
	cmd := exec.Command(cmds[0], cmds[1:]...)
	output, err := cmd.Output()
	return string(output), err
}
