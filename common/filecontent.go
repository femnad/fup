package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/internal"
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

func GetMvCmd(src, dst string) marecmd.Input {
	cmd := fmt.Sprintf("mv %s %s", src, dst)
	sudo := !IsHomePath(dst)

	return marecmd.Input{Command: cmd, Sudo: sudo}
}

func getChmodCmd(target string, mode int) marecmd.Input {
	octal := strconv.FormatInt(int64(mode), 8)
	cmd := fmt.Sprintf("chmod %s %s", octal, target)
	sudo := !IsHomePath(target)

	return marecmd.Input{Command: cmd, Sudo: sudo}
}

func Checksum(f string) (string, error) {
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

func getStatSum(f string) (statSum, error) {
	cmd := fmt.Sprintf("stat -c %%a %s", f)
	out, err := marecmd.RunFormatError(marecmd.Input{Command: cmd, Sudo: true})
	if err != nil {
		if strings.HasSuffix(strings.TrimSpace(out.Stderr), statNoExistsError) {
			return statSum{}, os.ErrNotExist
		}
		return statSum{}, err
	}

	mode, err := strconv.ParseUint(strings.TrimSpace(out.Stdout), 10, 32)
	if err != nil {
		return statSum{}, err
	}

	out, err = marecmd.RunFormatError(marecmd.Input{Command: fmt.Sprintf("sha256sum %s", f), Sudo: true})
	if err != nil {
		return statSum{}, err
	}

	sumFields := strings.Split(out.Stdout, "  ")
	if len(sumFields) != 2 {
		return statSum{}, fmt.Errorf("unexpected sha256sum output: %s", out.Stdout)
	}
	sum := sumFields[0]

	return statSum{mode: int(mode), sha256sum: sum}, nil
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

	return marecmd.RunNoOutput(marecmd.Input{Command: chownCmd, Sudo: true})
}

func ensureDir(dir string) error {
	if HasPerms(dir) {
		return os.MkdirAll(dir, 0o755)
	}

	internal.Log.Warningf("escalating privileges to create directory %s", dir)
	mkdirCmd := fmt.Sprintf("mkdir -p %s", dir)
	return marecmd.RunNoOutput(marecmd.Input{Command: mkdirCmd, Sudo: true})
}

func WriteContent(file ManagedFile) (bool, error) {
	var changed bool
	var dstSum string
	var srcSum string
	var noPermission bool
	var ss statSum
	dstExists := true
	target := internal.ExpandUser(file.Path)
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
	srcSum, err = Checksum(srcPath)
	if err != nil {
		return changed, err
	}
	defer os.Remove(srcPath)

	if dstExists {
		if noPermission {
			dstSum = ss.sha256sum
		} else {
			dstSum, err = Checksum(target)
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
		validateErr := marecmd.RunNoOutput(marecmd.Input{Command: validateCmd, Sudo: noPermission})
		if validateErr != nil {
			return false, validateErr
		}
	}

	dir, _ := path.Split(target)
	err = ensureDir(dir)
	if err != nil {
		return false, err
	}

	mv := GetMvCmd(srcPath, target)
	err = marecmd.RunNoOutput(mv)
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
		chmodCmd := getChmodCmd(target, mode)
		chmodErr := marecmd.RunNoOutput(chmodCmd)
		if chmodErr != nil {
			return changed, chmodErr
		}
	}

	if noPermission || !IsHomePath(target) {
		_, err = marecmd.RunFormatError(marecmd.Input{Command: fmt.Sprintf("chown %s:%s %s", rootUser, rootUser, target), Sudo: true})
		if err != nil {
			return changed, err
		}
	}

	err = chown(target, file.User, file.Group)
	if err != nil {
		return false, err
	}

	return true, nil
}
