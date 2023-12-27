package provision

import (
	"errors"
	"fmt"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

func ensureDirAbsent(dir entity.Dir) error {
	dirName := internal.ExpandUser(dir.Name)
	if err := internal.EnsureDirAbsent(dirName); err != nil {
		return fmt.Errorf("error removing directory %s: %v", dirName, err)
	}

	return nil
}

func ensureDirExist(dir entity.Dir) error {
	dirName := internal.ExpandUser(dir.Name)
	if err := internal.EnsureDirExists(dirName); err != nil {
		return fmt.Errorf("error creating directory %s: %v", dirName, err)
	}

	return nil
}

func doEnsureDirs(dirs []entity.Dir) error {
	var errs []error

	for _, dir := range dirs {
		var err error
		if dir.Absent {
			err = ensureDirAbsent(dir)
		} else {
			err = ensureDirExist(dir)
		}
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func ensureDirs(config base.Config) error {
	err := doEnsureDirs(config.EnsureDirs)
	if err != nil {
		internal.Log.Errorf("error ensuring dirs: %v", err)
		return err
	}

	return nil
}
