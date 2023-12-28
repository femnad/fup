package provision

import (
	"errors"
	"fmt"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

func ensureDirAbsent(dir string) error {
	dirName := internal.ExpandUser(dir)
	if err := internal.EnsureDirAbsent(dirName); err != nil {
		return fmt.Errorf("error removing directory %s: %v", dirName, err)
	}

	return nil
}

func ensureDirExist(dir string) error {
	dirName := internal.ExpandUser(dir)
	if err := internal.EnsureDirExists(dirName); err != nil {
		return fmt.Errorf("error creating directory %s: %v", dirName, err)
	}

	return nil
}

func doEnsureDirs(dirs []string, absent bool) []error {
	var errs []error

	for _, dir := range dirs {
		var err error
		if absent {
			err = ensureDirAbsent(dir)
		} else {
			err = ensureDirExist(dir)
		}
		errs = append(errs, err)
	}

	return errs
}

func ensureGroups(groups []entity.DirGroup) error {
	var groupErrs []error
	for _, group := range groups {
		errs := doEnsureDirs(group.Names, group.Absent)
		groupErrs = append(groupErrs, errs...)
	}

	return errors.Join(groupErrs...)
}

func ensureDirs(config entity.Config) error {
	err := ensureGroups(config.Dirs)
	if err != nil {
		internal.Log.Errorf("error ensuring dirs: %v", err)
		return err
	}

	return nil
}
