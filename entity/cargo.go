package entity

import (
	"fmt"

	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/settings"
)

type CargoPkg struct {
	unless.BasicUnlessable
	Bins    bool          `yaml:"bins"`
	Crate   string        `yaml:"name"`
	Unless  unless.Unless `yaml:"unless"`
	Version string        `yaml:"version"`
	When    string        `yaml:"when"`
}

func (c CargoPkg) DefaultVersionCmd() string {
	return fmt.Sprintf("%s --version", c.Name())
}

func (c CargoPkg) GetUnless() unless.Unless {
	return c.Unless
}

func (c CargoPkg) GetVersion(_ settings.Settings) (string, error) {
	return c.Version, nil
}

func (c CargoPkg) Name() string {
	return c.Crate
}

func (c CargoPkg) RunWhen() string {
	return c.When
}
