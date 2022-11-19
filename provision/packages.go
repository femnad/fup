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

func installPackages(spec base.PackageSpec) error {
	osId, err := precheck.GetOsId()
	if err != nil {
		return fmt.Errorf("error determining OS: %v", err)
	}

	installer, err := getInstaller(osId)
	if err != nil {
		return fmt.Errorf("cannot determine installer: %v", err)
	}

	pkgToInstall := mapset.NewSet[string]()
	for pattern, pkgs := range spec {
		matches := matchingPackages(osId, pattern, pkgs)
		for _, match := range matches {
			pkgToInstall.Add(match)
		}
	}

	return installer.Install(pkgToInstall)
}
