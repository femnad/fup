package base

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

func createSymlink(step Step, cfg Config) error {
	name := internal.ExpandUser(step.LinkSrc)
	target := ExpandSettings(cfg.Settings, step.LinkTarget)
	return common.Symlink(name, target)
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
	return common.RunCommandWithOutput(*cmd)
}

func runShellCmd(step Step, cfg Config) error {
	return common.RunShellCmd(step.Cmd, step.Sudo)
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
