package base

type Replacement struct {
	Old string `yaml:"old"`
	New string `yaml:"new"`
}

type LineInFile struct {
	Name    string        `yaml:"name"`
	File    string        `yaml:"file"`
	Replace []Replacement `yaml:"replace"`
	When    string        `yaml:"when"`
}

func (l LineInFile) RunWhen() string {
	return l.When
}
