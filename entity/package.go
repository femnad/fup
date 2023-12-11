package entity

type RemotePackage struct {
	// Some packages add their repos to OS repo config, causing conflicts if a previous version is set in the config.
	InstallOnce bool   `yaml:"install_once"`
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Url         string `yaml:"url"`
}
