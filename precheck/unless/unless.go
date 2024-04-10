package unless

import (
	"fmt"
	"os"
	"strings"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/run"
	"github.com/femnad/fup/settings"
	marecmd "github.com/femnad/mare/cmd"
)

type Unless struct {
	Cmd      string `yaml:"cmd"`
	ExitCode int    `yaml:"exit_code"`
	Post     string `yaml:"post"`
	Pwd      string `yaml:"pwd"`
	Shell    bool   `yaml:"shell"`
	Stat     string `yaml:"stat"`
}

func (u Unless) String() string {
	if u.Stat != "" {
		return fmt.Sprintf("ls %s", u.Stat)
	}

	s := u.Cmd
	if u.Post != "" {
		s += " | " + u.Post
	}
	return s
}

type Unlessable interface {
	DefaultVersionCmd() string
	GetUnless() Unless
	KeepUpToDate() bool
	LookupVersion(settings.Settings) (string, error)
	Name() string
}

type BasicUnlessable struct {
}

func (BasicUnlessable) KeepUpToDate() bool {
	return true
}

func doPostProcOutput(unless Unless, output string) (string, error) {
	postProcResult, err := internal.RunTemplateFn(output, unless.Post)
	if err != nil {
		return "", err
	}

	internal.Log.Debugf("postproc returned `%s` for `%s`", postProcResult, unless)
	return postProcResult, nil
}

func postProcOutput(unless Unless, output string) (string, error) {
	postProc := strings.TrimSpace(output)
	if unless.Post == "" {
		return postProc, nil
	}

	return doPostProcOutput(unless, postProc)
}

func shouldSkip(unlessable Unlessable, s settings.Settings) bool {
	var err error
	var out marecmd.Output
	unless := unlessable.GetUnless()
	unlessCmd := unless.Cmd

	if unlessCmd == "" {
		unlessCmd = unlessable.DefaultVersionCmd()
	}

	pwd := internal.ExpandUser(unless.Pwd)
	out, err = run.Cmd(s, marecmd.Input{Command: unlessCmd, Pwd: pwd, Shell: unless.Shell})

	if unless.ExitCode != 0 {
		internal.Log.Debugf("Command %s exited with code: %d, skip when: %d", unlessCmd, out.Code,
			unless.ExitCode)
		return out.Code == unless.ExitCode
	}

	if err != nil {
		internal.Log.Debugf("Command %s returned error: %v, output: %s", unlessCmd, err, out.Stderr)
		// Command wasn't successfully run, should not skip.
		return false
	}

	name := unlessable.Name()
	if !unlessable.KeepUpToDate() {
		internal.Log.Debugf("Not checking version for %s as it doesn't need to be kept up-to-date", name)
		return true
	}

	version, err := unlessable.LookupVersion(s)
	if err != nil {
		internal.Log.Errorf("Error determining desired version for %s: %v", name, err)
		return false
	}

	if version == "" {
		// No version specification, but command has succeeded so should skip the operation.
		internal.Log.Debugf("%s has no version specification, assuming operation should be skipped", name)
		return true
	}

	postProc, err := postProcOutput(unless, out.Stdout)
	if err != nil {
		internal.Log.Errorf("Error running postproc function for %s: %v", name, err)
		// Post processor function failed, best not to skip the operation.
		return false
	}

	if postProc != version {
		internal.Log.Debugf("%s: existing version `%s`, required version `%s`", name, postProc, version)
		return false
	}

	return true
}

func resolveStat(stat string, unlessable Unlessable, s settings.Settings) string {
	lookup := map[string]string{}
	version, err := unlessable.LookupVersion(s)
	if err != nil {
		internal.Log.Errorf("Error resolving stat %s: %v", stat, err)
		return stat
	}

	if version == "" {
		version = s.Versions[unlessable.Name()]
	}
	if version != "" {
		lookup["version"] = version
	}

	return settings.ExpandStringWithLookup(s, stat, lookup)
}

func sudoStat(target string) bool {
	internal.Log.Debugf("Trying to access %s with elevated privileges", target)

	statCmd := fmt.Sprintf("stat %s", target)
	out, cmdErr := marecmd.Run(marecmd.Input{Command: statCmd, Sudo: true})

	if strings.HasSuffix(strings.TrimSpace(out.Stderr), "No such file or directory") {
		return false
	} else if cmdErr == nil {
		return true
	}

	return false
}

func fileExists(target string) bool {
	internal.Log.Debugf("Checking existence of %s", target)

	_, err := os.Stat(target)
	if err == nil {
		return true
	} else if os.IsPermission(err) {
		return sudoStat(target)
	}

	return false
}

func ShouldSkip(unlessable Unlessable, s settings.Settings) bool {
	unless := unlessable.GetUnless()
	stat := unless.Stat

	if stat != "" {
		stat = resolveStat(stat, unlessable, s)
		return fileExists(stat)
	}

	if unless.Cmd == "" && unlessable.DefaultVersionCmd() == "" {
		// No stat or command checks, should not skip.
		return false
	}

	return shouldSkip(unlessable, s)
}
