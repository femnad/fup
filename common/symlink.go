package common

import (
	"fmt"
	"os"
	"path"

	"github.com/femnad/fup/internal"
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
	fmt.Println("hi")
	exists, update := shouldUpdateSymlink(symlinkName, symlinkTarget)
	if !update {
		internal.Log.Debugf("Symlink %s already exists", symlinkName)
		return nil
	}

	symlinkDir, _ := path.Split(symlinkName)
	err := os.MkdirAll(symlinkDir, dirMode)
	if err != nil {
		return fmt.Errorf("Error creating symlink dir %s: %v", symlinkDir, err)
	}

	internal.Log.Debugf("Creating symlink target=%s, name=%s", symlinkTarget, symlinkName)
	if exists {
		if err = os.Remove(symlinkName); err != nil {
			return fmt.Errorf("Error removing existing symlink %s: %v", symlinkName, err)
		}
	}

	err = os.Symlink(symlinkTarget, symlinkName)
	if err != nil {
		return fmt.Errorf("Error creating symlink target=%s, name=%s: %v", symlinkTarget, symlinkName, err)
	}

	return nil
}
