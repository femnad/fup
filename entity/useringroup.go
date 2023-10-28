package entity

type Group struct {
	Name   string `yaml:"name"`
	Ensure bool   `yaml:"ensure"`
	System bool   `yaml:"system"`
}
