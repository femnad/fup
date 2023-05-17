package common

import (
	"bufio"
	"fmt"
	"github.com/femnad/fup/internal"
	"os"
	"path"
)

const permissions = 0o600

func EnsureLinesInFile(file string, lines []string) error {
	itemSet := internal.SetFromList[string](lines)

	file = internal.ExpandUser(file)
	dir, _ := path.Split(file)
	if err := EnsureDir(dir); err != nil {
		return err
	}

	_, err := os.Stat(file)
	var fd *os.File
	if os.IsNotExist(err) {
		fd, err = os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, permissions)
		if err != nil {
			return err
		}
	} else if err == nil {
		fd, err = os.OpenFile(file, os.O_RDWR, permissions)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		if itemSet.Contains(line) {
			itemSet.Remove(line)
		}
	}

	itemSet.Each(func(key string) bool {
		line := fmt.Sprintf("%s\n", key)
		_, err = fd.WriteString(line)
		if err != nil {
			return true
		}
		return false
	})

	return err
}
