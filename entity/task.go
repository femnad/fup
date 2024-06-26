package entity

import (
	"fmt"
	"os"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/remote"
	"github.com/femnad/fup/run"
	"github.com/femnad/fup/settings"
	marecmd "github.com/femnad/mare/cmd"
)

func createSymlink(step Step, cfg Config) error {
	name := internal.ExpandUser(step.Src)
	target := ExpandSettings(cfg.Settings, step.Target)
	return common.Symlink(name, target)
}

func runCmd(step Step, cfg Config) error {
	var pwd string
	c := ExpandSettings(cfg.Settings, step.Cmd)
	if step.Pwd != "" {
		pwd = ExpandSettings(cfg.Settings, step.Pwd)
	}

	isRoot, err := internal.IsUserRoot()
	if err != nil {
		return err
	}

	sudo := !isRoot && step.Sudo
	_, err = run.Cmd(cfg.Settings, marecmd.Input{Command: c, Sudo: sudo, Pwd: pwd})
	return err
}

func runShellCmd(step Step, cfg Config) error {
	pwd := ExpandSettings(cfg.Settings, step.Pwd)
	cmd := ExpandSettings(cfg.Settings, step.Cmd)

	isRoot, err := internal.IsUserRoot()
	if err != nil {
		return err
	}

	sudo := !isRoot && step.Sudo
	_, err = run.Cmd(cfg.Settings, marecmd.Input{
		Command:  cmd,
		Pwd:      pwd,
		Shell:    true,
		ShellCmd: step.Shell,
		Sudo:     sudo,
	})
	return err
}

func runGitClone(step Step, cfg Config) error {
	path := step.Repo.Path
	if path == "" {
		path = cfg.Settings.CloneDir
	}
	path = ExpandSettings(cfg.Settings, path)
	return CloneUnderPath(step.Repo, path, cfg.Settings.CloneEnv)
}

func fileCmd(step Step, cfg Config) error {
	target := settings.ExpandString(cfg.Settings, step.Target)
	content := settings.ExpandString(cfg.Settings, step.Content)

	_, err := internal.WriteContent(internal.ManagedFile{
		Path:        target,
		Content:     content,
		ValidateCmd: step.Validate,
		Mode:        step.Mode,
	})
	return err
}

func download(step Step, cfg Config) error {
	url, path := step.Url, ExpandSettings(cfg.Settings, step.Target)
	internal.Log.Debugf("Downloading %s into %s", url, path)
	return remote.Download(url, path)
}

func rename(step Step, cfg Config) error {
	src := ExpandSettings(cfg.Settings, step.Src)
	target := ExpandSettings(cfg.Settings, step.Target)
	return os.Rename(src, target)
}

func getStepFunction(step Step) (func(Step, Config) error, error) {
	switch step.StepName {
	case "cmd":
		return runCmd, nil
	case "download":
		return download, nil
	case "file":
		return fileCmd, nil
	case "git":
		return runGitClone, nil
	case "rename":
		return rename, nil
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
	unless.BasicUnlessable
	// For cmd and shell
	Cmd string `yaml:"cmd"`
	// For file
	Content string `yaml:"content"`
	// For file
	Mode int `yaml:"mode"`
	// For cmd and shell
	Pwd  string `yaml:"pwd"`
	Repo Repo   `yaml:"repo"`
	// For cmd
	Sudo bool `yaml:"sudo"`
	// For link and rename
	Shell string `yaml:"shell"`
	// For rename and symlink
	Src      string `yaml:"src"`
	StepName string `yaml:"name"`
	// For download, file, link and rename
	Target string        `yaml:"target"`
	Unless unless.Unless `yaml:"unless"`
	// For file
	Validate string `yaml:"validate"`
	// For download
	Url string `yaml:"url"`
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

func (Step) LookupVersion(_ settings.Settings) (string, error) {
	return "", nil
}

func (s Step) Name() string {
	return s.StepName
}

func (Step) DefaultVersionCmd() string {
	return ""
}

type Task struct {
	unless.BasicUnlessable
	Desc   string        `yaml:"name"`
	Hint   string        `yaml:"hint"`
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

func (t Task) DefaultVersionCmd() string {
	return ""
}

func (t Task) GetUnless() unless.Unless {
	return t.Unless
}

func (t Task) LookupVersion(_ settings.Settings) (string, error) {
	return "", nil
}

func (t Task) Name() string {
	return t.Desc
}

func (t Task) Run(cfg Config) error {
	for _, step := range t.Steps {
		err := runStep(step, cfg)
		if err != nil {
			internal.Log.Errorf("error running task %s: %v", t.Name(), err)
			return err
		}
	}

	return nil
}
