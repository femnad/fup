package base

type PythonPkg struct {
	Pkg      string   `yaml:"name"`
	Reqs     []string `yaml:"reqs"`
	BinLinks []string `yaml:"link"`
}

func (p PythonPkg) GetUnless() Unless {
	return Unless{}
}

func (p PythonPkg) GetVersion() string {
	return ""
}

func (p PythonPkg) HasPostProc() bool {
	return false
}

func (p PythonPkg) Name() string {
	return p.Pkg
}
