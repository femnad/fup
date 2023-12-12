package entity

type FlatpakRemote struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

type FlatpakPkg struct {
	Launcher string `yaml:"launcher"`
	Name     string `yaml:"name"`
	Remote   string `yaml:"remote"`
}

type Flatpak struct {
	Remotes  []FlatpakRemote `yaml:"remote"`
	Packages []FlatpakPkg    `yaml:"pkg"`
}
