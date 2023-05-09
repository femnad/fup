package packages

import (
	"fmt"
	"os/user"
	"sort"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

const (
	rootUid = 0
)

type PkgManager interface {
	ListPkgsHeader() string
	PkgExec() string
	PkgNameSeparator() string
	RemoveCmd() string
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

	sudo := userId != rootUid

	cmd := strings.Join(cmds, " ")
	_, err = common.RunCmd(common.CmdIn{Command: cmd, Sudo: sudo})
	return err
}

type Installer struct {
	Pkg PkgManager
}

func setToSlice[T comparable](set mapset.Set[T]) []T {
	var items []T
	set.Each(func(t T) bool {
		items = append(items, t)
		return false
	})

	return items
}

func (i Installer) Install(desired mapset.Set[string]) error {
	available, err := i.installedPackages()
	if err != nil {
		return err
	}

	missing := desired.Difference(available)
	missingPkgs := setToSlice(missing)

	if len(missingPkgs) == 0 {
		return nil
	}

	sort.Strings(missingPkgs)
	internal.Log.Infof("Packages to install: %s", strings.Join(missingPkgs, " "))

	installCmd := []string{i.Pkg.PkgExec(), "install", "-y"}
	installCmd = append(installCmd, missingPkgs...)
	return maybeRunWithSudo(installCmd...)
}

func (i Installer) installedPackages() (mapset.Set[string], error) {
	listCmd := fmt.Sprintf("%s list --installed", i.Pkg.PkgExec())
	resp, err := common.RunCmd(common.CmdIn{Command: listCmd})
	if err != nil {
		return nil, err
	}

	installedPackages := mapset.NewSet[string]()
	lines := strings.Split(resp.Stdout, "\n")
	for _, line := range lines {
		if line == i.Pkg.ListPkgsHeader() {
			continue
		}
		fields := strings.Split(line, " ")
		if len(fields) == 0 {
			return nil, fmt.Errorf("unexpected package list line: %s", line)
		}

		pkgAndVers := fields[0]
		pkgFields := common.RightSplit(pkgAndVers, i.Pkg.PkgNameSeparator())
		if len(pkgFields) == 0 {
			return nil, fmt.Errorf("unexpected package field: %s", pkgFields)
		}

		packageName := pkgFields[0]
		installedPackages.Add(packageName)
	}

	return installedPackages, nil
}

func (i Installer) Remove(undesired mapset.Set[string]) error {
	available, err := i.installedPackages()
	if err != nil {
		return err
	}

	toRemove := available.Intersect(undesired)
	pkgToRemove := setToSlice(toRemove)

	if len(pkgToRemove) == 0 {
		return nil
	}

	sort.Strings(pkgToRemove)
	internal.Log.Infof("Packages to remove: %s", strings.Join(pkgToRemove, " "))

	removeCmd := []string{i.Pkg.PkgExec(), i.Pkg.RemoveCmd(), "-y"}
	removeCmd = append(removeCmd, pkgToRemove...)

	return maybeRunWithSudo(removeCmd...)
}
