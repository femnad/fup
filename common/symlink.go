package common

import (
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

func Symlink(symlinkName, symlinkTarget string) {
	exists, update := shouldUpdateSymlink(symlinkName, symlinkTarget)
	if !update {
		internal.Log.Debugf("Symlink %s already exists", symlinkName)
		return
	}

	symlinkDir, _ := path.Split(symlinkName)
	err := os.MkdirAll(symlinkDir, dirMode)
	if err != nil {
		internal.Log.Errorf("Error creating symlink dir %s: %v", symlinkDir, err)
		return
	}

	internal.Log.Debugf("Creating symlink target=%s, name=%s", symlinkTarget, symlinkName)
	if exists {
		if err = os.Remove(symlinkName); err != nil {
			internal.Log.Errorf("Error removing existing symlink %s: %v", symlinkName, err)
			return
		}
	}

	err = os.Symlink(symlinkTarget, symlinkName)
	if err != nil {
		internal.Log.Errorf("Error creating symlink target=%s, name=%s: %v", symlinkTarget, symlinkName, err)
	}
}
