package base

type Unit struct {
	Before      string            `yaml:"before"`
	Desc        string            `yaml:"desc"`
	Environment map[string]string `yaml:"env"`
	Exec        string            `yaml:"exec"`
	Options     map[string]string `yaml:"options"`
	WantedBy    string            `yaml:"wanted_by"`
}

type Service struct {
	DontEnable   bool   `yaml:"dont_enable"`
	DontStart    bool   `yaml:"dont_start"`
	DontTemplate bool   `yaml:"dont_template"`
	Disable      bool   `yaml:"disable"`
	Name         string `yaml:"name"`
	System       bool   `yaml:"system"`
	Stop         bool   `yaml:"stop"`
	Unit         Unit   `yaml:"unit"`
	When         string `yaml:"when"`
}

func (s Service) RunWhen() string {
	return s.When
}
