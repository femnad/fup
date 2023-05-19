package common

import (
	"os"
	"strings"
)

func EnsureDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0744)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func IsHomePath(path string) bool {
	home := os.Getenv("HOME")
	return strings.HasPrefix(path, home)
}

func HasPerms(targetPath string) bool {
	_, err := os.Stat(targetPath)
	if os.IsNotExist(err) {
		if strings.HasSuffix(targetPath, "/") {
			targetPath = targetPath[:len(targetPath)-1]
		}
		lastSlash := strings.LastIndex(targetPath, "/")
		parent := targetPath[:lastSlash]
		if parent == "" {
			return false
		}
		return HasPerms(parent)
	}

	return err == nil
}
