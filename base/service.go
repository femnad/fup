package base

type Unit struct {
	Exec string `yaml:"exec"`
	Desc string `yaml:"description"`
}

type Service struct {
	Desc string `yaml:"description"`
	Name string `yaml:"name"`
	Unit Unit   `yaml:"unit"`
}
