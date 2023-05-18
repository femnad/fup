package base

type Template struct {
	Src     string            `yaml:"src"`
	Dest    string            `yaml:"dest"`
	Mode    int               `yaml:"mode"`
	Context map[string]string `yaml:"context"`
	When    string            `yaml:"when"`
	User    string            `yaml:"owner"`
	Group   string            `yaml:"group"`
}

func (t Template) RunWhen() string {
	return t.When
}
