package entity

type DirGroup struct {
	Absent bool     `yaml:"absent"`
	Names  []string `yaml:"names"`
}
