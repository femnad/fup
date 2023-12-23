package entity

import (
	"fmt"

	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/precheck/unless"
)

type Binary struct {
	BinName string        `yaml:"name"`
	Dir     string        `yaml:"dir"`
	Url     string        `yaml:"url"`
	Unless  unless.Unless `yaml:"unless"`
	Version string        `yaml:"version"`
}

func (b Binary) DefaultVersionCmd() string {
	return fmt.Sprintf("%s --version", b.Name())
}

func (b Binary) GetUnless() unless.Unless {
	return b.Unless
}

func (b Binary) GetVersion(_ settings.Settings) (string, error) {
	return b.Version, nil
}

func (b Binary) Name() string {
	return b.BinName
}
