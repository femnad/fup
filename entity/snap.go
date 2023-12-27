package entity

type Snap struct {
	Absent  bool   `yaml:"absent"`
	Name    string `yaml:"name"`
	Classic bool   `yaml:"classic"`
}
