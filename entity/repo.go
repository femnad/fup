package entity

type Repo struct {
	Name      string            `yaml:"name"`
	Submodule bool              `yaml:"submodule"`
	Remotes   map[string]string `yaml:"remotes"`
}
