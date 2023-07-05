package entity

type RemotePackage struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Url     string `yaml:"url"`
}
