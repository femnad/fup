package base

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/femnad/fup/internal"
)

var shell = "sh"

func runCommand(cmd exec.Cmd) error {
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

func runCmd(step Step) error {
	cmds := strings.Split(step.Cmd, " ")
	var cmd *exec.Cmd
	if step.Sudo {
		cmd = exec.Command("sudo", cmds...)
	} else {
		cmd = exec.Command(cmds[0], cmds[1:]...)
	}
	return runCommand(*cmd)
}

func runShellCmd(step Step) error {
	var cmd *exec.Cmd
	if step.Sudo {
		cmd = exec.Command("sudo", []string{shell, "-c", step.Cmd}...)
	} else {
		cmd = exec.Command(shell, "-c", step.Cmd)
	}
	return runCommand(*cmd)
}

func getStepFunction(step Step) (func(Step) error, error) {
	switch step.Name {
	case "cmd":
		return runCmd, nil
	case "shell":
		return runShellCmd, nil
	case "":
		return nil, fmt.Errorf("no operation defined for step: %s", step)
	default:
		return nil, fmt.Errorf("unable to determine an operation for step: %s", step)
	}
}

type Step struct {
	Name string `yaml:"name"`
	Cmd  string `yaml:"cmd"`
	Sudo bool   `yaml:"sudo"`
}

func (s Step) String() string {
	return fmt.Sprintf("operation=%s, cmd=%s, sudo=%t", s.Name, s.Cmd, s.Sudo)
}

func (s Step) Run() error {
	fn, err := getStepFunction(s)
	if err != nil {
		return fmt.Errorf("error getting function for step: %v", err)
	}
	err = fn(s)
	if err != nil {
		return fmt.Errorf("error running function for step: %v", err)
	}
	return nil
}

type Task struct {
	Desc   string `yaml:"task"`
	Steps  []Step `yaml:"steps"`
	When   string `yaml:"when"`
	Unless Unless `yaml:"unless"`
}

func (t Task) RunWhen() string {
	return t.When
}

func (t Task) Run() {
	for _, step := range t.Steps {
		err := step.Run()
		if err != nil {
			internal.Log.Errorf("error running task %s: %v", t.Name, err)
			return
		}
	}
}

func (t Task) GetUnless() Unless {
	return t.Unless
}

func (t Task) GetVersion() string {
	return ""
}

func (t Task) HasPostProc() bool {
	return t.Unless.HasPostProc()
}

func (t Task) Name() string {
    return ""
}
