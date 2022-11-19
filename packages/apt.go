package packages

import mapset "github.com/deckarep/golang-set/v2"

type Apt struct {
	installer
}

func (a Apt) Install(packages mapset.Set[string]) error {
	installed, err := a.InstalledPackages()
	if err != nil {
		return err
	}

	return a.installer.Install("apt", installed, packages)
}

func (a Apt) InstalledPackages() (mapset.Set[string], error) {
	return a.installer.InstalledPackages("apt", "/")
}
