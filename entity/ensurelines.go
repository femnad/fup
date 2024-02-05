package entity

type Replacement struct {
	Absent bool   `yaml:"absent"`
	Old    string `yaml:"old"`
	New    string `yaml:"new"`
	Regex  bool   `yaml:"regex"`
}

type LineInFile struct {
	Content  []string      `yaml:"content"`
	File     string        `yaml:"file"`
	Name     string        `yaml:"name"`
	Replace  []Replacement `yaml:"replace"`
	RunAfter Step          `yaml:"run_after"`
	When     string        `yaml:"when"`
}

func (l LineInFile) RunWhen() string {
	return l.When
}
