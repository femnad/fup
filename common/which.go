package common

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func Which(exec string) (string, error) {
	if strings.HasPrefix(exec, "/") {
		_, err := os.Stat(exec)
		if err == nil {
			return exec, nil
		}
		return "", err
	}

	paths := os.Getenv("PATH")
	for _, pathComp := range strings.Split(paths, ":") {
		candidate := path.Join(pathComp, exec)
		_, err := os.Stat(candidate)
		if err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("unable to find absolute path for %s", exec)
}
