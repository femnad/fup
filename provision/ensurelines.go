package provision

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/when"
)

var (
	ensureFns = map[string]func(string, *os.File, entity.LineInFile) (ensureResult, error){
		"ensure":  ensure,
		"replace": replace,
	}
)

type ensureResult struct {
	changed bool
	new     bool
}

func ensure(file string, tmpFile *os.File, line entity.LineInFile) (result ensureResult, err error) {
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
		return
	}

	content.Each(func(l string) bool {
		_, err = tmpFile.WriteString(fmt.Sprintf("%s\n", l))
		return err != nil
	})

	if err != nil {
		return result, fmt.Errorf("error ensuring lines in file %s", file)
	}

	return ensureResult{changed: true, new: newFile}, nil
}

func replace(file string, tmpFile *os.File, line entity.LineInFile) (result ensureResult, err error) {
	srcFile, err := os.Open(file)
	if errors.Is(err, os.ErrNotExist) {
		return result, nil
	} else if err != nil {
		return
	}
	defer srcFile.Close()

	scanner := bufio.NewScanner(srcFile)
	scanner.Split(bufio.ScanLines)

	replacements := make(map[string]entity.Replacement)
	for _, replacement := range line.Replace {
		replacements[replacement.Old] = replacement
	}

	var changed bool
	for scanner.Scan() {
		var absent bool
		var found bool
		var oldLine string
		var newLine string
		l := scanner.Text()

		for _, needle := range replacements {
			var regex *regexp.Regexp
			absent = needle.Absent
			oldLine = needle.Old
			newLine = needle.New

			if needle.Regex {
				if l == newLine {
					found = true
					break
				}

				regex, err = regexp.Compile(oldLine)
				if err != nil {
					return result, err
				}

				if regex.MatchString(l) {
					changed = true
					found = true
					l = newLine
					break
				}
			} else if l == oldLine {
				changed = true
				found = true
				l = newLine
				break
			}
		}

		if found {
			delete(replacements, oldLine)
		}

		if absent && changed {
			continue
		}

		_, err = tmpFile.WriteString(l + "\n")
		if err != nil {
			return
		}
	}

	for _, replacement := range replacements {
		if replacement.Absent || !replacement.Ensure {
			continue
		}

		_, err = tmpFile.WriteString(replacement.New + "\n")
		if err != nil {
			return
		}
		changed = true
	}

	return ensureResult{changed: changed}, nil
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

	result, err := ensureFn(target, tmpFile, line)
	if err != nil {
		return err
	}

	tmpPath := tmpFile.Name()
	if !result.changed {
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

	err = internal.Move(tmpPath, target, result.new)
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
