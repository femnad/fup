package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"

	marecmd "github.com/femnad/mare/cmd"
)

const (
	// -1 means don't change user/group with os.Chown.
	defaultFileMode     = 0o644
	defaultId           = -1
	getentGroupDatabase = "group"
	getentSeparator     = ":"
	getentUserDatabase  = "passwd"
	rootUser            = "root"
	statNoExistsError   = "No such file or directory"
	tmpDir              = "/tmp"
)

type statSum struct {
	mode      int
	sha256sum string
}

type ManagedFile struct {
	Content     string
	Path        string
	Mode        int
	User        string
	Group       string
	ValidateCmd string
}

func checksum(f string) (string, error) {
	_, err := os.Stat(f)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	fd, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	_, err = io.Copy(h, fd)
	if err != nil {
		return "", err
	}

	sum := h.Sum(nil)
	return hex.EncodeToString(sum), nil
}

func chown(file, user, group string) error {
	isHomePath := IsHomePath(file)
	if user == "" && group == "" {
		if isHomePath {
			return nil
		}
		user = rootUser
		group = rootUser
	}

	userId, err := getent(user, getentUserDatabase)
	if err != nil {
		return err
	}

	groupId, err := getent(group, getentGroupDatabase)
	if err != nil {
		return err
	}

	if isHomePath {
		return os.Chown(file, userId, groupId)
	}

	chownCmd := "chown "
	if user != "" {
		chownCmd += user
	}
	if group != "" {
		chownCmd += ":" + user
	}
	chownCmd += " " + file

	return MaybeRunWithSudo(chownCmd)
}

func chmod(target string, mode int) error {
	octal := strconv.FormatInt(int64(mode), 8)
	chmodCmd := fmt.Sprintf("chmod %s %s", octal, target)
	return MaybeRunWithSudoForPath(chmodCmd, target)
}

func ensureDir(dir string) error {
	hasDirPerms, err := hasPerms(dir)
	if err != nil {
		return err
	}

	if hasDirPerms {
		return os.MkdirAll(dir, 0o755)
	}

	mkdirCmd := fmt.Sprintf("mkdir -p %s", dir)
	return MaybeRunWithSudo(mkdirCmd)
}

func getent(key, database string) (int, error) {
	if key == "" {
		return defaultId, nil
	}

	out, err := marecmd.RunFormatError(marecmd.Input{Command: fmt.Sprintf("getent %s %s", database, key)})
	if err != nil {
		return 0, err
	}

	getentOutput := out.Stdout
	getentFields := strings.Split(getentOutput, getentSeparator)
	if len(getentFields) < 2 {
		return 0, fmt.Errorf("unexpected getent output: %s", getentOutput)
	}

	id, err := strconv.ParseInt(getentFields[2], 10, 32)
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func getStatSum(f string) (statSum, error) {
	var s statSum
	isRoot, err := IsUserRoot()
	if err != nil {
		return s, err
	}

	cmd := fmt.Sprintf("stat -c %%a %s", f)
	out, err := marecmd.RunFormatError(marecmd.Input{Command: cmd, Sudo: !isRoot})
	if err != nil {
		if strings.HasSuffix(strings.TrimSpace(out.Stderr), statNoExistsError) {
			return s, os.ErrNotExist
		}
		return s, err
	}

	mode, err := strconv.ParseUint(strings.TrimSpace(out.Stdout), 10, 32)
	if err != nil {
		return s, err
	}

	out, err = marecmd.RunFormatError(marecmd.Input{Command: fmt.Sprintf("sha256sum %s", f), Sudo: !isRoot})
	if err != nil {
		return s, err
	}

	sumFields := strings.Split(out.Stdout, "  ")
	if len(sumFields) != 2 {
		return s, fmt.Errorf("unexpected sha256sum output: %s", out.Stdout)
	}
	sum := sumFields[0]

	return statSum{mode: int(mode), sha256sum: sum}, nil
}

func hasPerms(targetPath string) (bool, error) {
	fi, err := os.Stat(targetPath)
	if os.IsPermission(err) {
		return false, err
	} else if os.IsNotExist(err) {
		if strings.HasSuffix(targetPath, "/") {
			targetPath = targetPath[:len(targetPath)-1]
		}
		lastSlash := strings.LastIndex(targetPath, "/")
		if lastSlash < 0 {
			return false, fmt.Errorf("expected path %s to have at least one /", targetPath)
		}
		parent := targetPath[:lastSlash]
		if parent == "" {
			return false, nil
		}
		return hasPerms(parent)
	}

	stat, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("error determining owner of %s", targetPath)
	}

	userId, err := GetCurrentUserId()
	if err != nil {
		return false, err
	}

	return stat.Uid == uint32(userId), nil
}

func IsHomePath(path string) bool {
	home := os.Getenv("HOME")
	return strings.HasPrefix(path, home)
}

func WriteContent(file ManagedFile) (bool, error) {
	var changed bool
	var dstSum string
	var srcSum string
	var noPermission bool
	var ss statSum
	dstExists := true
	target := ExpandUser(file.Path)
	mode := file.Mode
	validateCmd := file.ValidateCmd

	_, err := os.Stat(target)
	if os.IsPermission(err) {
		noPermission = true
		ss, err = getStatSum(target)
		if os.IsNotExist(err) {
			dstExists = false
		} else if err != nil {
			return changed, err
		}
	} else if os.IsNotExist(err) {
		dstExists = false
	} else if err != nil {
		return changed, err
	}

	src, err := os.CreateTemp(tmpDir, "fup")
	if err != nil {
		return changed, err
	}
	defer src.Close()
	srcPath := src.Name()

	_, err = src.WriteString(file.Content)
	if err != nil {
		return changed, err
	}
	srcSum, err = checksum(srcPath)
	if err != nil {
		return changed, err
	}
	defer os.Remove(srcPath)

	if dstExists {
		if noPermission {
			dstSum = ss.sha256sum
		} else {
			dstSum, err = checksum(target)
			if err != nil {
				return changed, err
			}
		}
	}

	if dstExists && srcSum == dstSum {
		return false, nil
	}

	if validateCmd != "" {
		validateCmd = fmt.Sprintf("%s %s", validateCmd, srcPath)
		err = MaybeRunWithSudo(validateCmd)
		if err != nil {
			return false, err
		}
	}

	dir, _ := path.Split(target)
	err = ensureDir(dir)
	if err != nil {
		return changed, err
	}

	mv := fmt.Sprintf("mv %s %s", srcPath, target)

	err = MaybeRunWithSudoForPath(mv, target)
	if err != nil {
		return changed, err
	}

	if mode == 0 {
		mode = defaultFileMode
	}

	currentMode := defaultFileMode
	if noPermission {
		currentMode = ss.mode
	} else {
		fi, statErr := os.Stat(target)
		if statErr != nil {
			return false, fmt.Errorf("unexpected stat error for %s: %v", target, statErr)
		}
		currentMode = int(fi.Mode())
	}

	if currentMode != mode || !dstExists {
		err = chmod(target, mode)
		if err != nil {
			return changed, err
		}
	}

	if noPermission || !IsHomePath(target) {
		chownCmd := fmt.Sprintf("chown %s:%s %s", rootUser, rootUser, target)
		err = MaybeRunWithSudoForPath(chownCmd, target)
		return changed, err
	}

	err = chown(target, file.User, file.Group)
	if err != nil {
		return false, err
	}

	return true, nil
}
