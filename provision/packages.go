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

func getMatchingPackages(osId, pattern string, packages []string) []string {
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

func getInstaller(osId string) (packages.Os, error) {
	switch osId {
	case "fedora":
		return packages.Dnf{}, nil
	case "debian", "ubuntu":
		return packages.Apt{}, nil
	default:
		return nil, fmt.Errorf("no installer for OS ID %s", osId)
	}
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

	packagesToInstall := mapset.NewSet[string]()
	for pattern, desiredPackages := range spec {
		matches := getMatchingPackages(osId, pattern, desiredPackages)
		for _, match := range matches {
			packagesToInstall.Add(match)
		}
	}

	return installer.Install(packagesToInstall)
}
