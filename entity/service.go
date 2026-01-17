package entity

type Unit struct {
	Before      string            `yaml:"before"`
	Desc        string            `yaml:"desc"`
	Environment map[string]string `yaml:"env"`
	Exec        string            `yaml:"exec"`
	Options     map[string]string `yaml:"options"`
	Type        string            `yaml:"type"`
	WantedBy    string            `yaml:"wanted_by"`
}

type Timer struct {
	Calendar        string `yaml:"calendar"`
	Desc            string `yaml:"desc"`
	RandomizedDelay string `yaml:"randomized_delay"`
}

type Service struct {
	DontEnable   bool   `yaml:"dont_enable"`
	DontStart    bool   `yaml:"dont_start"`
	DontTemplate bool   `yaml:"dont_template"`
	Disable      bool   `yaml:"disable"`
	Kind         string `yaml:"kind"`
	Name         string `yaml:"name"`
	System       bool   `yaml:"system"`
	Stop         bool   `yaml:"stop"`
	Unit         *Unit  `yaml:"unit,omitempty"`
	Timer        *Timer `yaml:"timer,omitempty"`
	When         string `yaml:"when"`
}

func (s Service) RunWhen() string {
	return s.When
}
