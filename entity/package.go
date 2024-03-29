package entity

type RemotePackage struct {
	// Some packages add their repos to OS repo config, causing conflicts if a previous version is set in the config.
	InstallOnce bool   `yaml:"install_once"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	SkipScripts bool   `yaml:"skip_scripts"`
	Url         string `yaml:"url"`
}

type RemotePackageGroup struct {
	Pkgs []RemotePackage `yaml:"pkg"`
	When string          `yaml:"when"`
}

func (r RemotePackageGroup) RunWhen() string {
	return r.When
}

type PackageGroup struct {
	Absent bool     `yaml:"absent"`
	Pkgs   []string `yaml:"pkg"`
	When   string   `yaml:"when"`
}

func (p PackageGroup) RunWhen() string {
	return p.When
}
