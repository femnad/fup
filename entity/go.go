package entity

import (
	"fmt"

	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/settings"
)

type GoPkg struct {
	unless.BasicUnlessable
	Pkg     string        `yaml:"name"`
	Unless  unless.Unless `yaml:"unless"`
	Version string        `yaml:"version"`
	When    string        `yaml:"when"`
}

func (g GoPkg) DefaultVersionCmd() string {
	return fmt.Sprintf("%s version", g.Name())
}

func (g GoPkg) GetUnless() unless.Unless {
	return g.Unless
}

func (g GoPkg) GetVersion(_ settings.Settings) (string, error) {
	return g.Version, nil
}

func (g GoPkg) Name() string {
	return g.Pkg
}

func (g GoPkg) RunWhen() string {
	return g.When
}
