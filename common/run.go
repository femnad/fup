package common

import (
	"fmt"
	"github.com/femnad/fup/base/settings"
	marecmd "github.com/femnad/mare/cmd"
	"strings"
)

const (
	pathEnvKey    = "PATH"
	pathSeparator = ":"
)

func amendEnv(s settings.Settings, input marecmd.Input) marecmd.Input {
	path := strings.Join(s.EnsurePaths, pathSeparator)

	if input.Env == nil {
		input.Env = map[string]string{}
	}

	existingPathEnv, ok := input.Env[pathEnvKey]
	if ok {
		input.Env[pathEnvKey] = fmt.Sprintf("%s%s%s", existingPathEnv, pathSeparator, path)
	} else {
		input.Env[pathEnvKey] = path
	}

	return input
}

func RunCmd(s settings.Settings, input marecmd.Input) (marecmd.Output, error) {
	input = amendEnv(s, input)
	return marecmd.RunFormatError(input)
}

func RunCmdNoOutput(s settings.Settings, input marecmd.Input) error {
	input = amendEnv(s, input)
	return marecmd.RunNoOutput(input)
}
