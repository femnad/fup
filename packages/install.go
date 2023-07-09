package packages

import (
	"fmt"
	"os/user"
	"sort"
	"strconv"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

const (
	rootUid = 0
)

type PkgManager interface {
	ListPkgsHeader() string
	PkgExec() string
	PkgEnv() map[string]string
	PkgNameSeparator() string
	PreserveEnv() bool
	RemoveCmd() string
	RemoteInstall(urls []string) error
}

func isUserRoot() (bool, error) {
	currentUser, err := user.Current()
	if err != nil {
		return false, err
	}

	userId, err := strconv.ParseInt(currentUser.Uid, 10, 32)
	if err != nil {
		return false, err
	}

	return userId != rootUid, nil
}

type Installer struct {
	Pkg       PkgManager
	Installed mapset.Set[string]
}

func setToSlice[T comparable](set mapset.Set[T]) []T {
	var items []T
	set.Each(func(t T) bool {
		items = append(items, t)
		return false
	})

	return items
}

func (i Installer) maybeRunWithSudo(cmds ...string) error {
	sudo, err := isUserRoot()
	if err != nil {
		return err
	}

	cmdstr := strings.Join(cmds, " ")
	_, err = marecmd.RunFormatError(marecmd.Input{
		Command:         cmdstr,
		Sudo:            sudo,
		SudoPreserveEnv: i.Pkg.PreserveEnv(),
		Env:             i.Pkg.PkgEnv(),
	})
	return err
}

func (i Installer) Install(desired mapset.Set[string]) error {
	missing := desired.Difference(i.Installed)
	missingPkgs := setToSlice(missing)

	if len(missingPkgs) == 0 {
		return nil
	}

	sort.Strings(missingPkgs)
	internal.Log.Infof("Packages to install: %s", strings.Join(missingPkgs, " "))

	installCmd := []string{i.Pkg.PkgExec(), "install", "-y"}
	installCmd = append(installCmd, missingPkgs...)
	return i.maybeRunWithSudo(installCmd...)
}

func (i Installer) Version(pkg string) (string, error) {
	input := marecmd.Input{Command: fmt.Sprintf("%s info %s", i.Pkg.PkgExec(), pkg)}
	out, err := marecmd.RunFormatError(input)
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(out.Stdout, "\n") {
		if !strings.HasPrefix(line, "Version") {
			continue
		}
		fields := strings.Split(strings.TrimSpace(line), ": ")
		numFields := len(fields)
		if numFields != 2 {
			return "", fmt.Errorf("unexpected number of version fields in line %s", line)
		}
		return fields[numFields-1], nil
	}

	return "", err
}

func desiredPkgVersion(pkg entity.RemotePackage, s settings.Settings) string {
	version := pkg.Version
	if version == "" {
		version = s.Versions[pkg.Name]
	}

	return version
}

func (i Installer) RemoteInstall(desired mapset.Set[entity.RemotePackage], s settings.Settings) error {
	missing := mapset.NewSet[entity.RemotePackage]()

	var existingVersion string
	var vErr error
	desired.Each(func(pkg entity.RemotePackage) bool {
		if i.Installed.Contains(pkg.Name) {
			existingVersion, vErr = i.Version(pkg.Name)
			if vErr != nil {
				return true
			}

			if existingVersion != desiredPkgVersion(pkg, s) {
				missing.Add(pkg)
			}
		} else {
			missing.Add(pkg)
		}

		return false
	})

	if vErr != nil {
		return vErr
	}

	if missing.Equal(mapset.NewSet[entity.RemotePackage]()) {
		return nil
	}

	var urls []string
	missing.Each(func(pkg entity.RemotePackage) bool {
		version := desiredPkgVersion(pkg, s)
		url := settings.ExpandStringWithLookup(s, pkg.Url, map[string]string{"version": version})
		urls = append(urls, url)
		return false
	})

	sort.Strings(urls)
	internal.Log.Infof("Remote packages to install: %s", strings.Join(urls, " "))

	return i.Pkg.RemoteInstall(urls)
}

func (i Installer) InstalledPackages(pkg PkgManager) (mapset.Set[string], error) {
	listCmd := fmt.Sprintf("%s list --installed", pkg.PkgExec())
	resp, err := marecmd.RunFormatError(marecmd.Input{Command: listCmd})
	if err != nil {
		return nil, err
	}

	installedPackages := mapset.NewSet[string]()
	lines := strings.Split(resp.Stdout, "\n")
	for _, line := range lines {
		if line == pkg.ListPkgsHeader() {
			continue
		}
		fields := strings.Split(line, " ")
		if len(fields) == 0 {
			return nil, fmt.Errorf("unexpected package list line: %s", line)
		}

		pkgAndVers := fields[0]
		pkgFields := common.RightSplit(pkgAndVers, pkg.PkgNameSeparator())
		if len(pkgFields) == 0 {
			return nil, fmt.Errorf("unexpected package field: %s", pkgFields)
		}

		packageName := pkgFields[0]
		installedPackages.Add(packageName)
	}

	return installedPackages, nil
}

func (i Installer) Remove(undesired mapset.Set[string]) error {
	toRemove := i.Installed.Intersect(undesired)
	pkgToRemove := setToSlice(toRemove)

	if len(pkgToRemove) == 0 {
		return nil
	}

	sort.Strings(pkgToRemove)
	internal.Log.Infof("Packages to remove: %s", strings.Join(pkgToRemove, " "))

	removeCmd := []string{i.Pkg.PkgExec(), i.Pkg.RemoveCmd(), "-y"}
	removeCmd = append(removeCmd, pkgToRemove...)

	return i.maybeRunWithSudo(removeCmd...)
}
