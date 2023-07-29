package common

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/femnad/fup/internal"
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

func HasPerms(targetPath string) (bool, error) {
	fi, err := os.Stat(targetPath)
	if os.IsPermission(err) {
		return false, err
	} else if os.IsNotExist(err) {
		if strings.HasSuffix(targetPath, "/") {
			targetPath = targetPath[:len(targetPath)-1]
		}
		lastSlash := strings.LastIndex(targetPath, "/")
		if lastSlash < 0 {
			return false, fmt.Errorf("expected path %s to have at least one /", targetPath)
		}
		parent := targetPath[:lastSlash]
		if parent == "" {
			return false, nil
		}
		return HasPerms(parent)
	}

	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("error determining owner of %s", targetPath)
	}

	userId, err := internal.GetCurrentUserId()
	if err != nil {
		return false, err
	}

	return stat.Uid == uint32(userId), nil
}
