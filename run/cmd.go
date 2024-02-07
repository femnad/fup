package run

import (
	"fmt"
	"strings"

	"github.com/femnad/fup/settings"
	"github.com/femnad/mare"
	marecmd "github.com/femnad/mare/cmd"
)

const (
	pathEnvKey    = "PATH"
	pathSeparator = ":"
)

func amendEnv(s settings.Settings, input marecmd.Input) marecmd.Input {
	ensurePaths := mare.MapToString(s.EnsurePaths, func(s string) string {
		return mare.ExpandUser(s)
	})
	path := strings.Join(ensurePaths, pathSeparator)

	if input.Env == nil {
		input.Env = map[string]string{}
	}

	existingPathEnv, ok := input.Env[pathEnvKey]
	if ok {
		input.Env[pathEnvKey] = fmt.Sprintf("%s%s%s", existingPathEnv, pathSeparator, path)
	} else {
		input.Env[pathEnvKey] = path
	}

	for k, v := range s.EnsureEnv {
		input.Env[k] = mare.ExpandUser(v)
	}

	return input
}

func Cmd(s settings.Settings, input marecmd.Input) (marecmd.Output, error) {
	input = amendEnv(s, input)
	return marecmd.RunFmtErr(input)
}
