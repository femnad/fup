package base

import (
	"fmt"

	"github.com/femnad/fup/precheck/unless"
)

type GoPkg struct {
	Pkg     string        `yaml:"name"`
	Unless  unless.Unless `yaml:"unless"`
	Version string        `yaml:"version"`
}

func (g GoPkg) DefaultVersionCmd() string {
	return fmt.Sprintf("%s version", g.Name())
}

func (g GoPkg) GetUnless() unless.Unless {
	return g.Unless
}

func (g GoPkg) GetVersion() (string, error) {
	return g.Version, nil
}

func (g GoPkg) HasPostProc() bool {
	return false
}

func (g GoPkg) Name() string {
	return g.Pkg
}
