package common

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var shell = "sh"

func RunCommandWithOutput(cmd exec.Cmd) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return nil
	}

	stdoutStr := stdout.String()
	var output string
	if stdoutStr != "" {
		output += fmt.Sprintf("stdout: %s", stdoutStr)
	}
	stderrStr := stderr.String()
	if stderrStr != "" {
		if output != "" {
			output += ", "
		}
		output += fmt.Sprintf("stderr: %s", stderrStr)
	}
	return fmt.Errorf("error running command %s: %v => %s", cmd.String(), err, output)
}

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

func RunCmdExitCode(c string) int {
	cmds := strings.Split(c, " ")
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Run()
	return cmd.ProcessState.ExitCode()
}

func RunShellCmd(cmdstr string, sudo bool) error {
	var cmd *exec.Cmd
	if sudo {
		cmd = exec.Command("sudo", []string{shell, "-c", cmdstr}...)
	} else {
		cmd = exec.Command(shell, "-c", cmdstr)
	}
	return RunCommandWithOutput(*cmd)
}
