package provision

import (
	"errors"
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
	var installer packages.Installer
	var pkg packages.PkgManager

	switch osId {
	case "debian", "ubuntu":
		pkg = packages.Apt{}
	case "fedora":
		pkg = packages.Dnf{}
	default:
		return packages.Installer{}, fmt.Errorf("no installer for OS ID %s", osId)
	}

	installed, err := installer.InstalledPackages(pkg)
	if err != nil {
		return installer, err
	}

	return packages.Installer{
		Pkg:       pkg,
		Installed: installed,
	}, nil
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

type packager struct {
	determiner determiner
	installer  packages.Installer
}

func newPackager() (packager, error) {
	d, err := newDeterminer()
	if err != nil {
		return packager{}, err
	}

	i, err := d.installer()
	if err != nil {
		return packager{}, err
	}

	return packager{
		determiner: d,
		installer:  i,
	}, nil
}

func (p packager) installPackages(spec base.PackageSpec) error {
	pkgToInstall := p.determiner.matchingPkgs(spec)
	return p.installer.Install(pkgToInstall)
}

func (p packager) installRemotePackages(spec base.RemotePackageSpec, s settings.Settings) (bool, error) {
	pkgToInstall, err := p.determiner.matchingRemotePkgs(spec)
	if err != nil {
		return false, err
	}

	return p.installer.RemoteInstall(pkgToInstall, s)
}

func (p packager) removePackages(spec base.PackageSpec) error {
	pkgToInstall := p.determiner.matchingPkgs(spec)
	return p.installer.Remove(pkgToInstall)
}

func installPackages(p Provisioner) error {
	var pkgErrs []error

	var remoteUpdates bool
	remoteUpdates, err := p.Packager.installRemotePackages(p.Config.RemotePackages, p.Config.Settings)
	if err != nil {
		internal.Log.Errorf("error installing remote packages: %v", err)
		pkgErrs = append(pkgErrs, err)
	}

	if remoteUpdates {
		// Remote packages may have initialized repositories which may require updating the package database.
		if err = p.Packager.installer.Update(); err != nil {
			internal.Log.Errorf("error updating package database: %v", err)
			pkgErrs = append(pkgErrs, err)
		}
	}

	if err = p.Packager.installPackages(p.Config.Packages); err != nil {
		internal.Log.Errorf("error installing packages: %v", err)
		pkgErrs = append(pkgErrs, err)
	}

	return errors.Join(pkgErrs...)
}
