package provision

import (
	"bufio"
	"fmt"
	"os"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

func ensureLine(line base.LineInFile) error {
	if !when.ShouldRun(line) {
		internal.Log.Debugf("Skipping ensuring %s in %s due to condition %s", line.Replace, line.File, line.When)
		return nil
	}

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

	replacements := make(map[string]string)
	for _, replacement := range line.Replace {
		replacements[replacement.Old] = replacement.New
	}

	found := false
	for scanner.Scan() {
		var lineToWrite string
		l := scanner.Text()

		newLine, ok := replacements[l]
		if ok {
			lineToWrite = newLine
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

	mv, err := common.GetMvCmd(tmpPath, line.File)
	if err != nil {
		return err
	}

	out, err := marecmd.RunFormatError(mv)
	if err != nil {
		return fmt.Errorf("error renaming %s to %s: %v, output %s", tmpPath, line.File, err, out.Stderr)
	}

	return nil
}

func ensureLines(config base.Config) error {
	for _, line := range config.EnsureLines {
		err := ensureLine(line)
		if err != nil {
			internal.Log.Errorf("error ensuring lines in file: %v", err)
			return err
		}
	}

	return nil
}
