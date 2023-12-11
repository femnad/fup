package verify

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

func isDir(info os.FileInfo) bool {
	return info.IsDir()
}

func isRegularFile(info os.FileInfo) bool {
	return info.Mode().IsRegular()
}

func isExecutable(info os.FileInfo) bool {
	return common.IsExecutableFile(info)
}

func isSymlink(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}

var typeToVerifyFn = map[string]func(info os.FileInfo) bool{
	"dir":     isDir,
	"exec":    isExecutable,
	"file":    isRegularFile,
	"symlink": isSymlink,
}

func determineType(info os.FileInfo) string {
	if isDir(info) {
		return "dir"
	}

	if isSymlink(info) {
		return "symlink"
	}

	if isExecutable(info) {
		return "exec"
	}

	return "file"
}

func ensureCorrectDirEntry(entry DirEntry, fupConfig base.Config) error {
	path := internal.ExpandUser(entry.Path)
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	typ := entry.Type
	fn, ok := typeToVerifyFn[typ]
	if !ok {
		return fmt.Errorf("no verification function for dir entry type %s", typ)
	}

	if !fn(info) {
		actualType := determineType(info)
		return fmt.Errorf("%s has incorrect type, expected: %s, actual: %s", path, typ, actualType)
	}

	if typ == "symlink" {
		target := internal.ExpandUser(entry.Target)
		target = settings.ExpandString(fupConfig.Settings, target)
		if target == "" {
			return fmt.Errorf("symlink entry should have a target set")
		}

		dst, err := os.Readlink(path)
		if err != nil {
			return err
		}

		if dst != target {
			return fmt.Errorf("incorrect symlink target, expected: %s, actual: %s", target, dst)
		}
	}

	return nil
}

func readConfig(config string) (expect, error) {
	var e expect
	f, err := os.Open(config)
	if err != nil {
		return e, err
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&e)
	return e, err
}

func ensureExecutables(execs []string) []error {
	var errs []error
	for _, exec := range execs {
		_, err := common.Which(exec)
		errs = append(errs, err)
	}

	return errs
}

func ensureEntries(entries []DirEntry, config base.Config) []error {
	var errs []error
	for _, entry := range entries {
		err := ensureCorrectDirEntry(entry, config)
		errs = append(errs, err)
	}

	return errs
}

func ensureFileContent(content FileContent) error {
	path := internal.ExpandUser(content.Path)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if content.Content != string(bytes) {
		return fmt.Errorf("file %s doesn't have expected content", path)
	}

	return nil
}

func ensureFiles(files []FileContent) []error {
	var errs []error
	for _, content := range files {
		err := ensureFileContent(content)
		errs = append(errs, err)
	}

	return errs
}

func Verify(config string, fupConfig base.Config) error {
	e, err := readConfig(config)
	if err != nil {
		return err
	}

	execErrs := ensureExecutables(e.Executables)
	entryErrs := ensureEntries(e.DirEntries, fupConfig)
	fileErrs := ensureFiles(e.Files)

	var errs []error
	for _, errCol := range [][]error{execErrs, entryErrs, fileErrs} {
		errs = append(errs, errCol...)
	}

	return errors.Join(errs...)
}
