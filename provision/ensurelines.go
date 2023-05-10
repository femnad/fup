package provision

import (
	"bufio"
	"fmt"
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"os"
)

func ensureLine(line base.LineInFile) error {
	tmpFile, err := os.CreateTemp(tmpDir, "fup")
	if err != nil {
		return err
	}

	srcFile, err := os.Open(line.File)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	scanner := bufio.NewScanner(srcFile)
	scanner.Split(bufio.ScanLines)

	found := false
	for scanner.Scan() {
		var lineToWrite string
		l := scanner.Text()

		if l == line.Text {
			lineToWrite = line.Replace
			found = true
		} else {
			lineToWrite = l
		}

		_, err = tmpFile.WriteString(lineToWrite + "\n")
		if err != nil {
			return err
		}
	}

	tmpPath := tmpFile.Name()
	err = tmpFile.Close()
	if err != nil {
		return err
	}

	if !found {
		internal.Log.Debugf("Not modifying %s as line %s was not found", line.File, line.Replace)
		err = os.Remove(tmpPath)
		if err != nil {
			return err
		}
		return nil
	}

	mv := getMvCmd(tmpPath, line.File)
	out, err := common.RunCmd(mv)
	if err != nil {
		return fmt.Errorf("error renaming %s to %s: %v, output %s", tmpPath, line.File, err, out.Stderr)
	}

	return nil
}

func ensureLines(config base.Config) {
	for _, line := range config.EnsureLines {
		err := ensureLine(line)
		if err != nil {
			internal.Log.Errorf("error ensuring lines in file: %v", err)
		}
	}
}
