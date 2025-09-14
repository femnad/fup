package unless

import (
	"fmt"
	"log/slog"
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
	// For when the returned version is different from what needs to be replace the version part in URL.
	VersionOutput string `yaml:"version_output"`
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

	slog.Debug("postproc output`", "unless", unless, "result", postProcResult)
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
		slog.Debug("Command execution complete", "cmd", unlessCmd, "actual", out.Code,
			"expected", unless.ExitCode)
		return out.Code == unless.ExitCode
	}

	if err != nil {
		slog.Debug("Command returned error", "cmd", unlessCmd, "error", err, "stderr", out.Stderr)
		// Command wasn't successfully run, should not skip.
		return false
	}

	name := unlessable.Name()
	if !unlessable.KeepUpToDate() {
		slog.Debug("Not checking version for as it doesn't need to be kept up-to-date", "name", name)
		return true
	}

	version, err := unlessable.LookupVersion(s)
	if err != nil {
		slog.Error("Error determining desired version", "name", name, "error", err)
		return false
	}

	if version == "" {
		// No version specification, but command has succeeded so should skip the operation.
		slog.Debug("No version specification, assuming operation should be skipped", "name", name)
		return true
	}

	postProc, err := postProcOutput(unless, out.Stdout)
	if err != nil {
		slog.Error("Error running postproc function", "name", name, "error", err)
		// Post processor function failed, best not to skip the operation.
		return false
	}

	if unless.VersionOutput != "" {
		version = unless.VersionOutput
	}

	if postProc != version {
		slog.Debug("Actual and desired version mismatch", "name", name, "actual", postProc,
			"desired", version)
		return false
	}

	return true
}

func resolveStat(stat string, unlessable Unlessable, s settings.Settings) string {
	lookup := map[string]string{}
	version, err := unlessable.LookupVersion(s)
	if err != nil {
		slog.Error("Error resolving stat", "stat", stat, "error", err)
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
	slog.Debug("Trying access with elevated privileges", "target", target)

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
	slog.Debug("Checking file existence", "path", target)

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
