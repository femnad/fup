package base

type LineInFile struct {
	Name    string `yaml:"name"`
	File    string `yaml:"file"`
	Replace string `yaml:"replace"`
	Text    string `yaml:"text"`
	When    string `yaml:"when"`
}

func (l LineInFile) RunWhen() string {
	return l.When
}
