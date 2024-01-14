package provision

import (
	"bufio"
	"fmt"
	"github.com/femnad/fup/entity"
	"os"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

func ensureLine(config entity.Config, line entity.LineInFile) error {
	if !when.ShouldRun(line) {
		internal.Log.Debugf("Skipping changes to %s due to condition %s", line.File, line.When)
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
	removals := make(map[string]bool)
	for _, replacement := range line.Replace {
		if replacement.Absent {
			removals[replacement.Old] = true
			continue
		}

		replacements[replacement.Old] = replacement.New
	}

	changed := false
	for scanner.Scan() {
		var lineToWrite string
		l := scanner.Text()

		_, remove := removals[l]
		if remove {
			changed = true
			continue
		}

		newLine, ok := replacements[l]
		if ok {
			lineToWrite = newLine
			changed = true
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

	file := line.File
	if !changed {
		internal.Log.Debugf("Not modifying %s as no changes were found", file)
		err = os.Remove(tmpPath)
		if err != nil {
			return err
		}
		return nil
	}

	mv := fmt.Sprintf("mv %s %s", tmpPath, file)
	err = internal.MaybeRunWithSudoForPath(mv, file)
	if err != nil {
		return fmt.Errorf("error renaming %s to %s: %v", tmpPath, file, err)
	}

	fmt.Printf("%+v\n", line.RunAfter)

	executeAfter := line.RunAfter
	if executeAfter.Name() == "" {
		return nil
	}

	internal.Log.Debugf("Executing step %s after modifying %s", executeAfter.Name(), file)
	return executeAfter.Run(config)
}

func ensureLines(config entity.Config) error {
	for _, line := range config.EnsureLines {
		err := ensureLine(config, line)
		if err != nil {
			internal.Log.Errorf("error ensuring lines in file: %v", err)
			return err
		}
	}

	return nil
}
