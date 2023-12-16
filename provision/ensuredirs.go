package provision

import (
	"fmt"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
)

func doEnsureDirs(dirs []string) error {
	for _, dir := range dirs {
		dir = internal.ExpandUser(dir)
		if err := internal.EnsureDir(dir); err != nil {
			return fmt.Errorf("error ensuring directory %s: %v", dir, err)
		}
	}

	return nil
}

func ensureDirs(config base.Config) error {
	err := doEnsureDirs(config.EnsureDirs)
	if err != nil {
		internal.Log.Errorf("error ensuring dirs: %v", err)
		return err
	}

	return nil
}
