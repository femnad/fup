package unless

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/run"
)

var funcMap = template.FuncMap{
	"cut":     cut,
	"head":    head,
	"split":   split,
	"splitBy": splitBy,
}

type Unless struct {
	Cmd      string `yaml:"cmd"`
	ExitCode int    `yaml:"exit_code"`
	Post     string `yaml:"post"`
	Shell    bool   `yaml:"shell"`
	Stat     string `yaml:"stat"`
}

func (u Unless) HasPostProc() bool {
	return u.Post == "" || u.Cmd == ""
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
	GetVersion() string
	HasPostProc() bool
	Name() string
}

func absIndex(s string, i int) (int, error) {
	sLen := len(s)
	if i < 0 {
		i = sLen + i
	}
	if i < 0 || i >= sLen {
		return 0, fmt.Errorf("invalid index %d for string %s", i, s)
	}

	return i, nil
}

func cut(i int, s string) (string, error) {
	i, err := absIndex(s, i)
	if err != nil {
		return "", err
	}

	return s[i:], nil
}

func splitBy(delimiter string, i int, s string) (string, error) {
	fields := strings.Split(s, delimiter)
	numFields := len(fields)
	if i == -1 {
		i = numFields - 1
	}
	if i >= numFields {
		return "", fmt.Errorf("input %s has not have field with index %d when split by %s", s, i, delimiter)
	}
	return fields[i], nil
}

func head(i int, s string) (string, error) {
	return splitBy("\n", i, s)
}

func split(i int, s string) (string, error) {
	return splitBy(" ", i, s)
}

func applyProc(proc, output string) (string, error) {
	tmpl := template.New("post-proc").Funcs(funcMap)

	tmplTxt := fmt.Sprintf("{{ print `%s` | %s }}", output, proc)
	parsed, err := tmpl.Parse(tmplTxt)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	err = parsed.Execute(&out, context.TODO())
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func doPostProcOutput(unless Unless, output string) (string, error) {
	procs := strings.Split(unless.Post, "|")
	postOutput := output
	var err error

	for _, proc := range procs {
		postOutput, err = applyProc(proc, postOutput)
		if err != nil {
			return postOutput, err
		}
	}

	internal.Log.Debugf("postproc returned `%s` for `%s`", postOutput, unless)
	return postOutput, nil
}

func postProcOutput(unless Unless, output string) (string, error) {
	postProc := strings.TrimSpace(output)
	if unless.Post == "" {
		return postProc, nil
	}

	return doPostProcOutput(unless, postProc)
}

func getVersion(u Unlessable, s settings.Settings) string {
	version := u.GetVersion()
	if version != "" {
		return version
	}

	name := u.Name()
	if name == "" {
		return version
	}

	return s.Versions[name]
}

func shouldSkip(unlessable Unlessable, s settings.Settings) bool {
	var err error
	var out marecmd.Output
	unless := unlessable.GetUnless()
	unlessCmd := unless.Cmd
	if unlessCmd == "" {
		unlessCmd = unlessable.DefaultVersionCmd()
	}

	out, err = run.Cmd(s, marecmd.Input{Command: unlessCmd, Shell: unless.Shell})

	if unless.ExitCode != 0 {
		internal.Log.Debugf("Command %s exited with code: %d, skip when: %d", unlessCmd, out.Code, unless.ExitCode)
		return out.Code == unless.ExitCode
	}

	if err != nil {
		internal.Log.Debugf("Command %s returned error: %v, output: %s", unlessCmd, err, out.Stderr)
		// Command wasn't successfully run, should not skip.
		return false
	}

	version := getVersion(unlessable, s)
	if version == "" || unlessable.HasPostProc() {
		// No version specification or no post proc, but command has succeeded so should skip the operation.
		return true
	}

	postProc, err := postProcOutput(unless, out.Stdout)
	if err != nil {
		internal.Log.Errorf("Error running postproc function: %v", err)
		// Post processor function failed, best not to skip the operation.
		return false
	}

	vers := getVersion(unlessable, s)
	if postProc != vers {
		internal.Log.Debugf("Existing version `%s`, required version `%s`", postProc, vers)
		return false
	}

	return true
}

func resolveStat(stat string, unlessable Unlessable, s settings.Settings) string {
	lookup := map[string]string{}
	version := unlessable.GetVersion()

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
