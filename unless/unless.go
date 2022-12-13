package precheck

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

type Unlessable interface {
	GetUnless() base.Unless
	GetVersion() string
	HasPostProc() bool
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

func doPostProcOutput(unless base.Unless, output string) (string, error) {
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

func postProcOutput(unless base.Unless, output string) (string, error) {
	postProc := strings.TrimSpace(output)
	if unless.Post == "" {
		return postProc, nil
	}

	return doPostProcOutput(unless, postProc)
}

func shouldSkip(unlessable Unlessable) bool {
	unless := unlessable.GetUnless()
	cmd := unless.Cmd
	output, err := common.RunCmdGetStderr(cmd)
	if err != nil {
		internal.Log.Debugf("Command %s returned error: %v, output: %s", unless.Cmd, err, output)
		// Command wasn't successfully run, should not skip.
		return false
	}

	version := unlessable.GetVersion()
	if version == "" || unlessable.HasPostProc() {
		// No version specification or no post proc, but command has succeeded so should skip the operation.
		return true
	}

	postProc, err := postProcOutput(unless, output)
	if err != nil {
		internal.Log.Errorf("Error running postproc function: %v", err)
		// Post processor function failed, best not to skip the operation.
		return false
	}

	return postProc == unlessable.GetVersion()
}

func ShouldSkip(unlessable Unlessable) bool {
	unless := unlessable.GetUnless()
	stat := unless.Stat

	if stat != "" {
		stat = internal.ExpandUser(stat)
		internal.Log.Debugf("Checking existence of %s", stat)
		_, err := os.Stat(stat)
		return err == nil
	}

	if unless.Cmd == "" {
		// No stat or command checks, should not skip.
		return false
	}

	return shouldSkip(unlessable)
}
