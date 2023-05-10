package base

type Unit struct {
	Desc        string            `yaml:"desc"`
	Environment map[string]string `yaml:"env"`
	Exec        string            `yaml:"exec"`
}

type Service struct {
	DontEnable   bool   `yaml:"dont_enable"`
	DontStart    bool   `yaml:"dont_start"`
	DontTemplate bool   `yaml:"dont_template"`
	Disable      bool   `yaml:"disable"`
	Name         string `yaml:"name"`
	System       bool   `yaml:"system"`
	Unit         Unit   `yaml:"unit"`
	When         string `yaml:"when"`
}

func (s Service) RunWhen() string {
	return s.When
}
