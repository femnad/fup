package provision

import (
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/femnad/fup/internal"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/packages"
	"github.com/femnad/fup/precheck"
	"github.com/femnad/fup/precheck/when"
	"github.com/femnad/fup/settings"
)

func matchingPackages(group entity.PackageGroup) []string {
	if !when.ShouldRun(group) {
		return []string{}
	}

	return group.Pkgs
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
	osId, err := precheck.GetOSId()
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

func (d determiner) matchingPkg(spec entity.PackageSpec, install bool) mapset.Set[string] {
	pkgToInstall := mapset.NewSet[string]()
	for _, group := range spec {
		if group.Absent && install {
			continue
		} else if !group.Absent && !install {
			continue
		}
		matches := matchingPackages(group)
		for _, match := range matches {
			pkgToInstall.Add(match)
		}
	}

	return pkgToInstall
}

func (d determiner) matchingRemotePkg(spec entity.RemotePackageSpec) (mapset.Set[entity.RemotePackage], error) {
	pkgToInstall := mapset.NewSet[entity.RemotePackage]()

	for _, group := range spec {
		if !when.ShouldRun(group) {
			continue
		}

		for _, pkg := range group.Pkgs {
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

func (p packager) ensurePackages(spec entity.PackageSpec) error {
	pkgToInstall := p.determiner.matchingPkg(spec, true)
	installErr := p.installer.Install(pkgToInstall)

	pkgToRemove := p.determiner.matchingPkg(spec, false)
	removeErr := p.installer.Remove(pkgToRemove)

	return errors.Join(installErr, removeErr)
}

func (p packager) installRemotePackages(spec entity.RemotePackageSpec, s settings.Settings) error {
	pkgToInstall, err := p.determiner.matchingRemotePkg(spec)
	if err != nil {
		return err
	}

	return p.installer.RemoteInstall(pkgToInstall, s)
}

func installPackages(p Provisioner) error {
	var pkgErrs []error

	err := p.Packager.installRemotePackages(p.Config.RemotePackages, p.Config.Settings)
	if err != nil {
		internal.Logger.Error().Err(err).Msg("Error installing remote packages")
		pkgErrs = append(pkgErrs, err)
	}

	err = p.Packager.ensurePackages(p.Config.Packages)
	if err != nil {
		internal.Logger.Error().Err(err).Msg("Error installing packages")
		pkgErrs = append(pkgErrs, err)
	}

	return errors.Join(pkgErrs...)
}
