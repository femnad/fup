package base

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/femnad/fup/common"
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

func createSymlink(step Step, cfg Config) error {
	name := internal.ExpandUser(step.LinkSrc)
	target := ExpandSettings(cfg.Settings, step.LinkTarget)
	common.Symlink(name, target)
	return nil
}

func runCmd(step Step, cfg Config) error {
	c := internal.ExpandUser(step.Cmd)
	cmds := strings.Split(c, " ")
	var cmd *exec.Cmd
	if step.Sudo {
		cmd = exec.Command("sudo", cmds...)
	} else {
		cmd = exec.Command(cmds[0], cmds[1:]...)
	}
	return runCommand(*cmd)
}

func runShellCmd(step Step, cfg Config) error {
	var cmd *exec.Cmd
	if step.Sudo {
		cmd = exec.Command("sudo", []string{shell, "-c", step.Cmd}...)
	} else {
		cmd = exec.Command(shell, "-c", step.Cmd)
	}
	return runCommand(*cmd)
}

func runGitClone(step Step, cfg Config) error {
	path := ExpandSettings(cfg.Settings, step.Dir)
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	opt := git.CloneOptions{
		URL: step.Repo,
	}
	_, err = git.PlainClone(path, false, &opt)
	return err
}

func getStepFunction(step Step) (func(Step, Config) error, error) {
	switch step.Name {
	case "cmd":
		return runCmd, nil
	case "git":
		return runGitClone, nil
	case "shell":
		return runShellCmd, nil
	case "symlink":
		return createSymlink, nil
	case "":
		return nil, fmt.Errorf("no operation defined for step: %s", step)
	default:
		return nil, fmt.Errorf("unable to determine an operation for step: %s", step)
	}
}

type Step struct {
	Cmd        string `yaml:"cmd"`
	Dir        string `yaml:"dir"`
	LinkSrc    string `yaml:"link_src"`
	LinkTarget string `yaml:"link_target"`
	Name       string `yaml:"name"`
	Repo       string `yaml:"repo"`
	Sudo       bool   `yaml:"sudo"`
}

func (s Step) String() string {
	return fmt.Sprintf("operation=%s, cmd=%s, sudo=%t", s.Name, s.Cmd, s.Sudo)
}

func (s Step) Run(cfg Config) error {
	fn, err := getStepFunction(s)
	if err != nil {
		return fmt.Errorf("error getting function for step: %v", err)
	}
	err = fn(s, cfg)
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

func (t Task) Run(cfg Config) {
	for _, step := range t.Steps {
		err := step.Run(cfg)
		if err != nil {
			internal.Log.Errorf("error running task %s: %v", t.Name(), err)
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
	return t.Desc
}
