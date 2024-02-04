package internal

import (
	"fmt"
	"os"
)

func EnsureDirAbsent(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func EnsureDirExists(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(dir, 0744)
	if err == nil {
		return nil
	} else if !os.IsPermission(err) {
		return err
	}

	err = MaybeRunWithSudo(fmt.Sprintf("mkdir -p %s", dir))
	if err != nil {
		return err
	}

	return chown(dir, rootUser, rootUser)
}
