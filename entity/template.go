package entity

type Template struct {
	Content    string            `yaml:"content"`
	Context    map[string]string `yaml:"context"`
	Dest       string            `yaml:"dest"`
	ExpandUser bool              `yaml:"expand_user"`
	Group      string            `yaml:"group"`
	Mode       int               `yaml:"mode"`
	RunAfter   []Step            `yaml:"run_after"`
	Src        string            `yaml:"src"`
	User       string            `yaml:"owner"`
	When       string            `yaml:"when"`
}

func (t Template) RunWhen() string {
	return t.When
}
