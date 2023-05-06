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
	Name         string `yaml:"name"`
	Unit         Unit   `yaml:"unit"`
}
