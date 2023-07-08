package unless

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/internal"
)

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
	GetUnless() Unless
	GetVersion() string
	HasPostProc() bool
	Name() string
}

func processString(fnName, separator, s string, i int, procFn func([]string, int) string) (string, error) {
	tokens := strings.Split(s, separator)
	lenTokens := len(tokens)

	if i > lenTokens {
		return "", fmt.Errorf("invalid %s index for input %s and index %d", fnName, s, i)
	}

	if i < 0 {
		iAbs := int(math.Abs(float64(i)))
		if iAbs > lenTokens-1 {
			return "", fmt.Errorf("invalid negative %s index for input %s and index %d", fnName, s, i)
		}
		i = lenTokens - iAbs
	}

	return procFn(tokens, i), nil
}

func delimitAndReturn(fnName, separator, s string, i int) (string, error) {
	processed, err := processString(fnName, separator, s, i, func(tokens []string, index int) string {
		return tokens[i]
	})
	if err != nil {
		return "", err
	}

	return processed, nil
}

func Cut(s string, i int) (string, error) {
	processed, err := processString("Cut", "", s, i, func(tokens []string, index int) string {
		lenTokens := len(tokens)
		if i > 0 {
			return strings.Join(tokens[i:lenTokens], "")
		}

		return strings.Join(tokens[:index], "")
	})
	if err != nil {
		return "", err
	}

	return processed, nil
}

func Head(s string, i int) (string, error) {
	return delimitAndReturn("Head", "\n", s, i)
}

func Split(s string, i int) (string, error) {
	return delimitAndReturn("Split", " ", s, i)
}

func SplitByDash(s string, i int) (string, error) {
	return delimitAndReturn("SplitByDash", "-", s, i)
}

func getPostProcFn(op string) (func(string, int) (string, error), error) {
	switch op {
	case "cut":
		return Cut, nil
	case "head":
		return Head, nil
	case "split":
		return Split, nil
	case "split-":
		return SplitByDash, nil
	default:
		return nil, fmt.Errorf("error locating post processing function for %s", op)
	}
}

func applyProc(proc, output string) (string, error) {
	proc = strings.TrimSpace(proc)
	postOutput := output
	fnInvocation := strings.Split(proc, " ")
	if len(fnInvocation) != 2 {
		return postOutput, fmt.Errorf("error parsing postproc functions args for %s", proc)
	}
	fnName := fnInvocation[0]

	fnArg, err := strconv.Atoi(fnInvocation[1])
	if err != nil {
		return postOutput, fmt.Errorf("error converting %s to index: %v", fnInvocation[1], err)
	}

	fn, err := getPostProcFn(fnName)
	if err != nil {
		return postOutput, fmt.Errorf("error getting postproc function for %s: %v", fnName, err)
	}

	postOutput, err = fn(postOutput, fnArg)
	if err != nil {
		return postOutput, fmt.Errorf("error running postproc function %s: %v", fnName, err)
	}

	return postOutput, nil
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

func shouldSkip(unlessable Unlessable, settings settings.Settings) bool {
	var err error
	var out marecmd.Output
	unless := unlessable.GetUnless()
	unlessCmd := unless.Cmd

	out, err = marecmd.Run(marecmd.Input{Command: unlessCmd, Shell: unless.Shell})

	if unless.ExitCode != 0 {
		internal.Log.Debugf("Command %s exited with code: %d, skip when: %d", unlessCmd, out.Code, unless.ExitCode)
		return out.Code == unless.ExitCode
	}

	if err != nil {
		internal.Log.Debugf("Command %s returned error: %v, output: %s", unlessCmd, err, out.Stderr)
		// Command wasn't successfully run, should not skip.
		return false
	}

	version := getVersion(unlessable, settings)
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

	vers := getVersion(unlessable, settings)
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

	if unless.Cmd == "" {
		// No stat or command checks, should not skip.
		return false
	}

	return shouldSkip(unlessable, s)
}
