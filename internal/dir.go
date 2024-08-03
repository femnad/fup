package internal

import (
	"fmt"
	"os"
)

func EnsureFileAbsent(file string) error {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	err = os.Remove(file)
	if os.IsPermission(err) {
		return MaybeRunWithSudo(fmt.Sprintf("rm %s", file))
	}

	return nil
}

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
	if dir == "" {
		return nil
	}

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

	return Chown(dir, rootUser, rootUser)
}
