package base

import "github.com/femnad/fup/precheck/unless"

type PythonPkg struct {
	Pkg      string        `yaml:"name"`
	Reqs     []string      `yaml:"reqs"`
	BinLinks []string      `yaml:"link"`
	Unless   unless.Unless `yaml:"unless"`
}

func (p PythonPkg) GetUnless() unless.Unless {
	return p.Unless
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
