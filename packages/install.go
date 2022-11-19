package packages

import (
	"fmt"
	"os/exec"
	"os/user"
	"sort"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

const (
	listPackagesHeader = "Installed Packages"
	rootUid            = 0
)

type Os interface {
	Install(mapset.Set[string]) error
	InstalledPackages() (mapset.Set[string], error)
}

func maybeRunWithSudo(cmds ...string) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	userId, err := strconv.ParseInt(currentUser.Uid, 10, 32)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	if userId == rootUid {
		cmd = exec.Command(cmds[0], cmds[1:]...)
	} else {
		cmd = exec.Command("sudo", cmds...)
	}

	return cmd.Run()
}

type installer struct {
}

func (i installer) Install(pkg string, available, desired mapset.Set[string]) error {
	missing := desired.Difference(available)
	var missingList []string
	missing.Each(func(p string) bool {
		missingList = append(missingList, p)
		return false
	})

	if len(missingList) == 0 {
		return nil
	}

	sort.Strings(missingList)
	internal.Log.Noticef("Packages to install: %s", strings.Join(missingList, " "))

	installCmd := []string{pkg, "install", "-y"}
	installCmd = append(installCmd, missingList...)
	return maybeRunWithSudo(installCmd...)
}

func (i installer) InstalledPackages(pkg, versionSeparator string) (mapset.Set[string], error) {
	output, err := common.RunCmd(fmt.Sprintf("%s list --installed", pkg))
	if err != nil {
		return nil, err
	}

	installedPackages := mapset.NewSet[string]()
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == listPackagesHeader {
			continue
		}
		fields := strings.Split(line, " ")
		if len(fields) == 0 {
			return nil, fmt.Errorf("unexpected package list line: %s", line)
		}
		packageAndVersion := fields[0]
		packageFields := strings.Split(packageAndVersion, versionSeparator)
		if len(packageFields) == 0 {
			return nil, fmt.Errorf("unexpected package field: %s", packageFields)
		}
		packageName := packageFields[0]
		installedPackages.Add(packageName)
	}

	return installedPackages, nil
}
