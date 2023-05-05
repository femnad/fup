package provision

import (
	"fmt"
	"os"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
)

func removeDirs(dirs []string) error {
	for _, dir := range dirs {
		dir = internal.ExpandUser(dir)
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		err = os.Remove(dir)
		if err != nil {
			return fmt.Errorf("error removing directory %s: %v", dir, err)
		}
	}

	return nil
}

func removeUnwantedDirs(config base.Config) {
	err := removeDirs(config.UnwantedDirs)
	if err != nil {
		internal.Log.Errorf("error removing unwanted dirs: %v", err)
	}
}
