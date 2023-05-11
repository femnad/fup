package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	defaultFileMode   = 0644
	statNoExistsError = "No such file or directory"
	tmpDir            = "/tmp"
)

type statSum struct {
	mode      uint32
	sha256sum string
}

func GetMvCmd(src, dst string) CmdIn {
	cmd := fmt.Sprintf("mv %s %s", src, dst)
	home := os.Getenv("HOME")
	sudo := !strings.HasPrefix(dst, home)

	return CmdIn{Command: cmd, Sudo: sudo}
}

func getChmodCmd(target string, mode os.FileMode) CmdIn {
	cmd := fmt.Sprintf("chmod %d %s", mode, target)
	home := os.Getenv("HOME")
	sudo := !strings.HasPrefix(target, home)

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

	return statSum{mode: uint32(mode), sha256sum: sum}, nil
}

func WriteContent(target, content, validate string, mode os.FileMode) error {
	var dstSum string
	var srcSum string
	var noPermission bool
	var ss statSum
	dstExists := true

	fi, err := os.Stat(target)
	if os.IsPermission(err) {
		noPermission = true
		ss, err = getStatSum(target)
		if os.IsNotExist(err) {
			dstExists = false
		} else if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		dstExists = false
	} else if err != nil {
		return err
	}

	src, err := os.CreateTemp(tmpDir, "fup")
	if err != nil {
		return err
	}
	defer src.Close()
	srcPath := src.Name()

	_, err = src.WriteString(content)
	if err != nil {
		return err
	}
	srcSum, err = Checksum(srcPath)
	if err != nil {
		return err
	}
	defer os.Remove(srcSum)

	if dstExists {
		if noPermission {
			dstSum = ss.sha256sum
		} else {
			dstSum, err = Checksum(target)
			if err != nil {
				return err
			}
		}
	}

	if dstExists && srcSum == dstSum {
		return nil
	}

	if validate != "" {
		validateCmd := fmt.Sprintf("%s %s", validate, srcPath)
		out, validateErr := RunCmd(CmdIn{Command: validateCmd, Sudo: noPermission})
		if validateErr != nil {
			return fmt.Errorf("error running validate command %s, output %s", validateCmd, strings.TrimSpace(out.Stderr))
		}
	}

	mv := GetMvCmd(srcPath, target)
	_, err = RunCmd(mv)
	if err != nil {
		return err
	}

	if mode != 0 {
		mode = defaultFileMode
	}

	var targetMode os.FileMode
	if noPermission {
		targetMode = os.FileMode(ss.mode)
	} else {
		targetMode = fi.Mode()
	}

	if targetMode != mode {
		chmodCmd := getChmodCmd(target, mode)
		_, err = RunCmd(chmodCmd)
		if err != nil {
			return err
		}
	}

	if noPermission {
		_, err = RunCmd(CmdIn{Command: fmt.Sprintf("chown 0:0 %s", target), Sudo: true})
		if err != nil {
			return err
		}
	}

	return err
}
