package common

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var shell = "sh"

type CmdOut struct {
	Code   int
	Stdout string
	Stderr string
}

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

func RunCmd(command string) (CmdOut, error) {
	cmds := strings.Split(command, " ")
	cmd := exec.Command(cmds[0], cmds[1:]...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return CmdOut{Stdout: stdout.String(), Stderr: stderr.String(), Code: cmd.ProcessState.ExitCode()}, err
}

func RunMaybeSudo(c string, sudo bool) (CmdOut, error) {
	if sudo {
		c = "sudo " + c
	}
	return RunCmd(c)
}

func runCmdGetOutput(command string, runInShell bool) (string, error) {
	var b bytes.Buffer
	cmds := strings.Split(command, " ")

	var cmd *exec.Cmd
	if runInShell {
		cmd = exec.Command(shell, "-c", command)
	} else {
		cmd = exec.Command(cmds[0], cmds[1:]...)
	}
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Run()

	return b.String(), err
}

func RunShellGetOutput(command string) (string, error) {
	return runCmdGetOutput(command, true)
}

func RunCmdGetStderr(command string) (string, error) {
	return runCmdGetOutput(command, false)
}

func RunShellCmd(cmdstr, pwd string, sudo bool) error {
	var cmd *exec.Cmd
	if sudo {
		cmd = exec.Command("sudo", []string{shell, "-c", cmdstr}...)
	} else {
		cmd = exec.Command(shell, "-c", cmdstr)
	}
	if pwd != "" {
		cmd.Dir = pwd
	}
	return RunCommandWithOutput(*cmd)
}
