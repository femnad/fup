package common

import (
	"fmt"
	"log/slog"
	"os"
	"path"
)

const (
	dirMode = 0755
)

func shouldUpdateSymlink(name, target string) (bool, bool) {
	_, err := os.Lstat(name)
	if err != nil {
		return false, true
	}

	currLink, err := os.Readlink(name)
	if err != nil {
		return true, true
	}

	return true, currLink != target
}

func Symlink(symlinkName, symlinkTarget string) error {
	if symlinkName == "" || symlinkTarget == "" {
		return fmt.Errorf("symlink name and target must be both set")
	}

	exists, update := shouldUpdateSymlink(symlinkName, symlinkTarget)
	if !update {
		slog.Debug("Symlink already exists", "name", symlinkName)
		return nil
	}

	symlinkDir, _ := path.Split(symlinkName)
	err := os.MkdirAll(symlinkDir, dirMode)
	if err != nil {
		return fmt.Errorf("error creating symlink dir %s: %v", symlinkDir, err)
	}

	slog.Debug("Creating symlink", "target", symlinkTarget, "name", symlinkName)
	if exists {
		if err = os.Remove(symlinkName); err != nil {
			return fmt.Errorf("error removing existing symlink %s: %v", symlinkName, err)
		}
	}

	_, err = os.Stat(symlinkTarget)
	if err != nil {
		return fmt.Errorf("error stat-ing symlink target %s: %v", symlinkTarget, err)
	}

	err = os.Symlink(symlinkTarget, symlinkName)
	if err != nil {
		return fmt.Errorf("error creating symlink target=%s, name=%s: %v", symlinkTarget, symlinkName, err)
	}

	return nil
}
