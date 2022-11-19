package packages

import (
	mapset "github.com/deckarep/golang-set/v2"
)

type Dnf struct {
	installer
}

func (f Dnf) Install(packages mapset.Set[string]) error {
	installed, err := f.InstalledPackages()
	if err != nil {
		return err
	}

	return f.installer.Install("dnf", installed, packages)
}

func (f Dnf) InstalledPackages() (mapset.Set[string], error) {
	return f.installer.InstalledPackages("dnf", ".")
}
