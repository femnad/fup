package base

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-git/go-git/v5"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
)

const (
	defaultFileMode = 0644
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
	if step.Pwd != "" {
		cmd.Dir = ExpandSettings(cfg.Settings, step.Pwd)
	}
	return common.RunCommandWithOutput(*cmd)
}

func runShellCmd(step Step, _ Config) error {
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

func fileCmd(step Step, _ Config) error {
	path := os.ExpandEnv(step.Path)
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	mode := defaultFileMode
	if step.Mode != 0 {
		mode = step.Mode
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, os.FileMode(mode))
	if err != nil {
		return err
	}

	_, err = f.WriteString(step.Content)
	return err
}

func getStepFunction(step Step) (func(Step, Config) error, error) {
	switch step.StepName {
	case "cmd":
		return runCmd, nil
	case "file":
		return fileCmd, nil
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
	Cmd        string        `yaml:"cmd"`
	Content    string        `yaml:"content"`
	Dir        string        `yaml:"dir"`
	LinkSrc    string        `yaml:"link_src"`
	LinkTarget string        `yaml:"link_target"`
	Mode       int           `yaml:"mode"`
	Path       string        `yaml:"path"`
	Pwd        string        `yaml:"pwd"`
	Repo       string        `yaml:"repo"`
	StepName   string        `yaml:"name"`
	Sudo       bool          `yaml:"sudo"`
	Unless     unless.Unless `yaml:"unless"`
}

func (s Step) String() string {
	return fmt.Sprintf("operation=%s, cmd=%s, sudo=%t", s.Name(), s.Cmd, s.Sudo)
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

func (s Step) GetUnless() unless.Unless {
	return s.Unless
}

func (Step) GetVersion() string {
	return ""
}

func (Step) HasPostProc() bool {
	return false
}

func (Step) Name() string {
	return ""
}

type Task struct {
	Desc   string        `yaml:"task"`
	Steps  []Step        `yaml:"steps"`
	When   string        `yaml:"when"`
	Unless unless.Unless `yaml:"unless"`
}

func (t Task) RunWhen() string {
	return t.When
}

func runStep(step Step, cfg Config) error {
	if unless.ShouldSkip(step, cfg.Settings) {
		internal.Log.Debugf("skipping step %s due to precheck", step.StepName)
		return nil
	}

	err := step.Run(cfg)
	if err != nil {
		return fmt.Errorf("error running step %s: %v", step.StepName, err)
	}

	return nil
}

func (t Task) Run(cfg Config) {
	for _, step := range t.Steps {
		err := runStep(step, cfg)
		if err != nil {
			internal.Log.Errorf("error running task %s: %v", t.Name(), err)
			return
		}
	}
}

func (t Task) GetUnless() unless.Unless {
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
