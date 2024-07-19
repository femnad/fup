package entity

type Archive struct {
	Files  []string `yaml:"files"`
	Target string   `yaml:"target"`
	URL    string   `yaml:"url"`
}
