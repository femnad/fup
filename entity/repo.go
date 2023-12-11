package entity

type Repo struct {
	Branch    string            `yaml:"branch"`
	Name      string            `yaml:"name"`
	Path      string            `yaml:"path"`
	Remotes   map[string]string `yaml:"remotes"`
	Submodule bool              `yaml:"submodule"`
	Tag       string            `yaml:"tag"`
	Update    bool              `yaml:"update"`
}
