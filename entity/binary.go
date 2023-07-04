package entity

import "github.com/femnad/fup/precheck/unless"

type Binary struct {
	BinName string        `yaml:"name"`
	Url     string        `yaml:"url"`
	Unless  unless.Unless `yaml:"unless"`
	Version string        `yaml:"version"`
}

func (b Binary) GetUnless() unless.Unless {
	return b.Unless
}

func (b Binary) GetVersion() string {
	return b.Version
}

func (b Binary) HasPostProc() bool {
	return b.Unless.HasPostProc()
}

func (b Binary) Name() string {
	return b.BinName
}
