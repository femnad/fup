package precheck

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
)

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

func getPostProcFn(op string) (func(string, int) (string, error), error) {
	switch op {
	case "cut":
		return Cut, nil
	case "head":
		return Head, nil
	case "split":
		return Split, nil
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

func postProcOutput(unless base.Unless, output string) (string, error) {
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

func shouldSkip(archive base.Archive) bool {
	unless := archive.Unless

	cmds := strings.Split(unless.Cmd, " ")
	cmd := exec.Command(cmds[0], cmds[1:]...)
	output, err := cmd.Output()
	if err != nil {
		// Command wasn't successfully run, should not skip.
		return false
	}

	if unless.Post == "" {
		// No post processor configuration but command has succeeded so should skip the operation.
		return true
	}

	postProc := strings.TrimSpace(string(output))
	postProc, err = postProcOutput(unless, postProc)
	if err != nil {
		internal.Log.Errorf("Error running postproc function: %v", err)
		// Post processor function failed, best not to skip the operation.
		return false
	}

	return postProc == archive.Version
}

func expandStat(archive base.Archive, settings base.Settings) string {
	return os.Expand(archive.Unless.Stat, func(s string) string {
		if s == "extract_dir" {
			extractDir := settings.ExtractDir
			return internal.ExpandUser(extractDir)
		}
		if s == "version" {
			return archive.Version
		}
		return s
	})
}

func ShouldSkip(archive base.Archive, settings base.Settings) bool {
	stat := expandStat(archive, settings)

	if stat != "" {
		internal.Log.Debugf("Checking existence of %s", stat)
		_, err := os.Stat(stat)
		return err == nil
	}

	if archive.Unless.Cmd == "" {
		// No stat or command checks, should not skip.
		return false
	}

	return shouldSkip(archive)
}
