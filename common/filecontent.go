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

func GetMvCmd(src, dst string) CmdIn {
	cmd := fmt.Sprintf("mv %s %s", src, dst)
	sudo := !IsHomePath(dst)

	return CmdIn{Command: cmd, Sudo: sudo}
}

func getChmodCmd(target string, mode int) CmdIn {
	octal := strconv.FormatInt(int64(mode), 8)
	cmd := fmt.Sprintf("chmod %s %s", octal, target)
	sudo := !IsHomePath(target)

	return CmdIn{Command: cmd, Sudo: sudo}
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
	out, err := RunCmd(CmdIn{Command: cmd, Sudo: true})
	if err != nil {
		if strings.HasSuffix(strings.TrimSpace(out.Stderr), statNoExistsError) {
			return statSum{}, os.ErrNotExist
		}
		return statSum{}, fmt.Errorf("error running command %s, output %s, error %v", cmd, out.Stderr, err)
	}
	mode, err := strconv.ParseUint(strings.TrimSpace(out.Stdout), 10, 32)
	if err != nil {
		return statSum{}, err
	}

	out, err = RunCmd(CmdIn{Command: fmt.Sprintf("sha256sum %s", f), Sudo: true})
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

	out, err := RunCmd(CmdIn{Command: fmt.Sprintf("getent %s %s", database, key)})
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

	out, err := RunCmd(CmdIn{Command: chownCmd, Sudo: true})
	if err != nil {
		return fmt.Errorf("error changing owner of %s, output %s: %v", file, out.Stderr, err)
	}

	return nil
}

func ensureDir(dir string) error {
	if HasPerms(dir) {
		return os.MkdirAll(dir, 0o755)
	}

	internal.Log.Warningf("escalating privileges to create directory %s", dir)
	mkdirCmd := fmt.Sprintf("mkdir -p %s", dir)
	out, err := RunCmd(CmdIn{Command: mkdirCmd, Sudo: true})
	if err != nil {
		return fmt.Errorf("error running command %s, output %s: %v", mkdirCmd, out.Stderr, err)
	}

	return nil
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
		out, validateErr := RunCmd(CmdIn{Command: validateCmd, Sudo: noPermission})
		if validateErr != nil {
			return changed, fmt.Errorf("error running validate command %s, output %s", validateCmd, strings.TrimSpace(out.Stderr))
		}
	}

	dir, _ := path.Split(target)
	err = ensureDir(dir)
	if err != nil {
		return false, err
	}

	mv := GetMvCmd(srcPath, target)
	out, err := RunCmd(mv)
	if err != nil {
		return changed, fmt.Errorf("error running mv command: %s, output %s: %v", mv.Command, out.Stderr, err)
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
		_, chmodErr := RunCmd(chmodCmd)
		if chmodErr != nil {
			return changed, chmodErr
		}
	}

	if noPermission || !IsHomePath(target) {
		_, err = RunCmd(CmdIn{Command: fmt.Sprintf("chown %s:%s %s", rootUser, rootUser, target), Sudo: true})
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
