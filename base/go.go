package base

import "github.com/femnad/fup/precheck/unless"

type GoPkg struct {
	Pkg     string        `yaml:"name"`
	Unless  unless.Unless `yaml:"unless"`
	Version string        `yaml:"version"`
}

func (g GoPkg) GetUnless() unless.Unless {
	return g.Unless
}

func (g GoPkg) GetVersion() string {
	return g.Version
}

func (g GoPkg) HasPostProc() bool {
	return false
}

func (g GoPkg) Name() string {
	return g.Pkg
}
