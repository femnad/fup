package entity

import (
	"fmt"

	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/settings"
)

type PythonPkg struct {
	unless.BasicUnlessable
	BinLinks      []string          `yaml:"link"`
	Library       bool              `yaml:"library"`
	Pkg           string            `yaml:"name"`
	Reqs          []string          `yaml:"reqs"`
	Unless        unless.Unless     `yaml:"unless"`
	Version       string            `yaml:"version"`
	VersionLookup VersionLookupSpec `yaml:"version_lookup"`
}

func (p PythonPkg) GetVersion() string {
	return p.Version
}

func (p PythonPkg) GetVersionLookup() VersionLookupSpec {
	return p.VersionLookup
}

func (p PythonPkg) GetLookupID() string {
	return p.Name()
}

func (p PythonPkg) DefaultVersionCmd() string {
	return fmt.Sprintf("%s -V", p.Name())
}

func (p PythonPkg) GetUnless() unless.Unless {
	return p.Unless
}

func (p PythonPkg) LookupVersion(s settings.Settings) (string, error) {
	return getVersion(p, s)
}

func (p PythonPkg) Name() string {
	return p.Pkg
}
