package entity

import (
	"fmt"

	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/settings"
)

type PythonPkg struct {
	unless.BasicUnlessable
	Pkg      string        `yaml:"name"`
	Reqs     []string      `yaml:"reqs"`
	BinLinks []string      `yaml:"link"`
	Unless   unless.Unless `yaml:"unless"`
}

func (p PythonPkg) DefaultVersionCmd() string {
	return fmt.Sprintf("%s -V", p.Name())
}

func (p PythonPkg) GetUnless() unless.Unless {
	return p.Unless
}

func (p PythonPkg) GetVersion(s settings.Settings) (string, error) {
	return "", nil
}

func (p PythonPkg) Name() string {
	return p.Pkg
}
