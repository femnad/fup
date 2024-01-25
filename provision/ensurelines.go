package provision

import (
	"bufio"
	"fmt"
	"os"
	"path"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

var (
	ensureFns = map[string]func(string, *os.File, entity.LineInFile) (bool, error){
		"ensure":  ensure,
		"replace": replace,
	}
)

func ensure(file string, tmpFile *os.File, line entity.LineInFile) (changed bool, err error) {
	content := mapset.NewSet[string]()
	for _, l := range line.Content {
		content.Add(l)
	}

	var newFile bool
	srcFile, err := os.Open(file)
	if os.IsNotExist(err) {
		newFile = true
	} else if err != nil {
		return
	}
	defer srcFile.Close()

	if !newFile {
		scanner := bufio.NewScanner(srcFile)
		for scanner.Scan() {
			l := scanner.Text()
			if content.Contains(l) {
				content.Remove(l)
			}
		}
	}

	if content.Cardinality() == 0 {
		return false, nil
	}

	content.Each(func(l string) bool {
		_, err = tmpFile.WriteString(fmt.Sprintf("%s\n", l))
		return err != nil
	})

	if err != nil {
		return false, fmt.Errorf("error ensuring lines in file %s", file)
	}

	return true, nil
}

func replace(file string, tmpFile *os.File, line entity.LineInFile) (changed bool, err error) {
	srcFile, err := os.Open(file)
	if err != nil {
		return
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
			return
		}
	}

	return changed, nil
}

func ensureLine(config entity.Config, line entity.LineInFile) error {
	target := internal.ExpandUser(line.File)
	if !when.ShouldRun(line) {
		internal.Log.Debugf("Skipping changes to %s due to condition %s", target, line.When)
		return nil
	}

	tmpFile, err := os.CreateTemp(tmpDir, "fup")
	if err != nil {
		return err
	}

	ensureFn, ok := ensureFns[line.Name]
	if !ok {
		return fmt.Errorf("no method for %s'ing a line", line.Name)
	}

	changed, err := ensureFn(target, tmpFile, line)
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	if !changed {
		internal.Log.Debugf("Not modifying %s as no changes were found", target)
		err = os.Remove(tmpPath)
		if err != nil {
			return err
		}
		return nil
	}

	targetDir, _ := path.Split(target)
	err = ensureDirExist(targetDir)
	if err != nil {
		return err
	}

	mv := fmt.Sprintf("mv %s %s", tmpPath, target)
	err = internal.MaybeRunWithSudoForPath(mv, target)
	if err != nil {
		return fmt.Errorf("error renaming %s to %s: %v", tmpPath, target, err)
	}

	executeAfter := line.RunAfter
	if executeAfter.Name() == "" {
		return nil
	}

	internal.Log.Debugf("Executing step %s after modifying %s", executeAfter.Name(), target)
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
