package verify

type DirEntry struct {
	Path   string `yaml:"path"`
	Target string `yaml:"target"`
	Type   string `yaml:"type"`
}

type FileContent struct {
	Content string `yaml:"content"`
	Path    string `yaml:"path"`
}

type expect struct {
	Executables []string      `yaml:"executables"`
	DirEntries  []DirEntry    `yaml:"paths"`
	Files       []FileContent `yaml:"file_content"`
}
