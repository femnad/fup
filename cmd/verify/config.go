package verify

type DirEntry struct {
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
	Type   string `yaml:"type"`
}

type expect struct {
	Executables []string   `yaml:"executables"`
	DirEntries  []DirEntry `yaml:"paths"`
}
