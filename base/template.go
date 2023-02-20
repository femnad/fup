package base

type Template struct {
	Src     string            `yaml:"src"`
	Dest    string            `yaml:"dest"`
	Mode    int               `yaml:"mode"`
	Context map[string]string `yaml:"context"`
	When    string            `yaml:"when"`
}

func (t Template) RunWhen() string {
	return t.When
}
