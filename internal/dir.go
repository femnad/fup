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
	if os.IsPermission(err) {
		return MaybeRunWithSudo(fmt.Sprintf("mkdir %s", dir))
	}
	return err
}
