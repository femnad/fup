package provision

import (
	"fmt"
	"regexp"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/packages"
	"github.com/femnad/fup/precheck"
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

func (d determiner) matchingRemotePkgs(spec base.RemotePackageSpec) (mapset.Set[entity.RemotePackage], error) {
	pkgToInstall := mapset.NewSet[entity.RemotePackage]()

	for pattern, pkgs := range spec {
		match, err := regexp.MatchString(pattern, d.osId)
		if err != nil {
			return pkgToInstall, err
		}

		if !match {
			continue
		}

		for _, pkg := range pkgs {
			pkgToInstall.Add(pkg)
		}
	}

	return pkgToInstall, nil
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

func installRemotePackages(spec base.RemotePackageSpec, s settings.Settings) error {
	d, err := newDeterminer()
	if err != nil {
		return err
	}

	i, err := d.installer()
	if err != nil {
		return err
	}

	pkgToInstall, err := d.matchingRemotePkgs(spec)
	if err != nil {
		return err
	}

	return i.RemoteInstall(pkgToInstall, s)
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
