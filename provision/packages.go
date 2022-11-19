package provision

import (
	"fmt"
	"regexp"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/packages"
	precheck "github.com/femnad/fup/unless"
)

func matchingPackages(osId, pattern string, packages []string) []string {
	match, err := regexp.MatchString(pattern, osId)
	if err != nil {
		internal.Log.Errorf("Error matching pattern %s: %v", pattern, err)
		return []string{}
	}

	if !match {
		return []string{}
	}

	return packages
}

func getInstaller(osId string) (packages.Installer, error) {
	installer := packages.Installer{}

	switch osId {
	case "debian", "ubuntu":
		installer.Pkg = packages.Apt{}
	case "fedora":
		installer.Pkg = packages.Dnf{}
	default:
		return installer, fmt.Errorf("no installer for OS ID %s", osId)
	}

	return installer, nil
}

type determiner struct {
	osId string
}

func newDeterminer() (determiner, error) {
	var d determiner
	osId, err := precheck.GetOsId()
	if err != nil {
		return d, fmt.Errorf("error determining OS: %v", err)
	}

	d.osId = osId
	return d, nil
}

func (d determiner) installer() (packages.Installer, error) {
	installer, err := getInstaller(d.osId)
	if err != nil {
		return installer, fmt.Errorf("cannot determine installer: %v", err)
	}

	return installer, nil
}

func (d determiner) matchingPkgs(spec base.PackageSpec) mapset.Set[string] {
	pkgToInstall := mapset.NewSet[string]()
	for pattern, pkgs := range spec {
		matches := matchingPackages(d.osId, pattern, pkgs)
		for _, match := range matches {
			pkgToInstall.Add(match)
		}
	}

	return pkgToInstall
}

func installPackages(spec base.PackageSpec) error {
	d, err := newDeterminer()
	if err != nil {
		return err
	}

	i, err := d.installer()
	if err != nil {
		return err
	}

	pkgToInstall := d.matchingPkgs(spec)
	return i.Install(pkgToInstall)
}

func removePackages(spec base.PackageSpec) error {
	d, err := newDeterminer()
	if err != nil {
		return err
	}

	i, err := d.installer()
	if err != nil {
		return err
	}

	pkgToInstall := d.matchingPkgs(spec)
	return i.Remove(pkgToInstall)
}
